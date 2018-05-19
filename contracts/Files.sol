pragma solidity 0.4.23;

/**
    This contract is used to act as a central, immmutable record of the files stored in our system, allowing independent audits of our data storage
*/

contract FileRepository {


    /*
        Whenever a file is uploaded, the release date will be checked, if the  processing
        pin request will persist longer than the current one, update release date
    */
    struct FileUpload {
        string ipfsHash; // the actual content hash of the file, if a directory it will be the parent directory hash
        uint256 releaseDate; 
        mapping (address => bool) uploaders;
    }

    mapping (bytes32 => FileUpload) public uploads;
    
    modifier onlyPaymentProcessor() {
        require(msg.sender == address(0)); // placeholder
        _;
    }

    function addUpload(
        address _uploader,
        string _ipfsHash,
        uint256 _releaseDate)
        public
        onlyPaymentProcessor
        returns (bool)
    {
        uploads[keccak256(_ipfsHash)].uploaders[_uploader] = true;
        if (uploads[keccak256(_ipfsHash)].releaseDate < _releaseDate) {
            uploads[keccak256(_ipfsHash)].releaseDate = _releaseDate;
        }
        // event placeholder
        return true;
    }
}