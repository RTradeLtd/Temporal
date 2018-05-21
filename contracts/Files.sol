pragma solidity 0.4.23;

/**
    This contract is used to act as a central, immmutable record of the files stored in our system, allowing independent audits of our data storage

current limitations:
    date calculate will be off, this is temporarily.
        when removing files from our system we will infer to an internal record
        this will be changed in subsequent releases
*/

import "./Math/SafeMath.sol";

contract FileRepository {

    using SafeMath for uint256;

    address public paymentProcessor;

    struct CidStruct{
        bytes32 hashedCID;
        uint256 numberOfTimesUploaded;
        uint256 removalDate;
        mapping (address => bool) uploaders;
        address[] uploaderArray;
    }

    mapping (bytes32 => CidStruct) public cids;

    modifier onlyPaymentProcessor() {
        require(msg.sender == paymentProcessor);
        _;
    }

    function addUploaderForCid(
        address _uploader,
        bytes32 _hashedCID,
        uint256 _retentionPeriodInMonths)
        public
        onlyPaymentProcessor
        returns (bool)
    {
        uint256 numDays = _retentionPeriodInMonths.mul(30);
        uint256 newRemovalDate = now.add((numDays.mul(1 days)));
        if (newRemovalDate > cids[_hashedCID].removalDate) {
            cids[_hashedCID].removalDate = newRemovalDate;
        }
        if (!cids[_hashedCID].uploaders[_uploader]) {
            cids[_hashedCID].uploaders[_uploader] = true;
            cids[_hashedCID].uploaderArray.push(_uploader);
        }
        cids[_hashedCID].numberOfTimesUploaded = cids[_hashedCID].numberOfTimesUploaded.add(1);
        return true;
    }
}