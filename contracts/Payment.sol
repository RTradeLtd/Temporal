pragma solidity 0.4.23;

/**
 Used by a user to pay for a file upload
 */


interface UserInterface {
    function paymentLockFunds(address _user, uint256 _amount) external returns (bool);
}

contract Payment {

    UserInterface public utI;

    struct PaymentStruct {
        uint256 date;
        uint256 numRTC;
        uint256 fileRetentionDurationInMonths;
        string fileHash; // if directory, refers to parent directory hash
    }

    mapping (bytes32 => PaymentStruct) payments;

    function payForUpload(
        uint256 _retentionPeriodInMonths,
        uint256 _paymentAmount,
        string _fileHash)
        public
        returns (bool)
    {
        PaymentStruct memory ps;
        ps.date = now;
        ps.numRTC = _paymentAmount;
        ps.fileRetentionDurationInMonths = _retentionPeriodInMonths;
        ps.fileHash = _fileHash;
        payments[keccak256(msg.sender, _fileHash, now)] = ps;
        require(utI.paymentLockFunds(msg.sender, _paymentAmount));
        return true;
    }
}