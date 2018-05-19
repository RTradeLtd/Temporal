pragma solidity 0.4.23;

/**
 Used by a user to pay for a file upload
 */


interface UserInterface {
    function paymentLockFunds(address _user, uint256 _amount) external returns (bool);
}

contract Payment {

    bytes private prefix = "\x19Ethereum Signed Message:\n32";

    UserInterface public utI;

    struct PaymentStruct {
        uint256 date;
        uint256 numRTC;
        uint256 fileRetentionDurationInMonths;
        bytes32 encryptedIpfsHash; // if directory, refers to parent directory hash
    }

    mapping (bytes32 => PaymentStruct) payments;

    function payForUpload(
        uint256 _retentionPeriodInMonths,
        uint256 _paymentAmount,
        bytes32 _encryptedIpfsHash)
        public
        returns (bool)
    {
        PaymentStruct memory ps;
        ps.date = now;
        ps.numRTC = _paymentAmount;
        ps.fileRetentionDurationInMonths = _retentionPeriodInMonths;
        ps.encryptedIpfsHash = _encryptedIpfsHash;
        payments[keccak256(msg.sender, _encryptedIpfsHash, now)] = ps;
        require(utI.paymentLockFunds(msg.sender, _paymentAmount));
        return true;
    }

    function constructPreimage(
        address _addr,
        bytes32 _encryptedFileHash,
        uint256 _durationInMonths,
        uint256 _chargeAmount)
        public
        pure
        returns (bytes32)
    {
        return keccak256(_addr, _encryptedFileHash, _durationInMonths, _chargeAmount);
    }

    function constructH(
        bytes32 _preimage)
        public
        view
        returns (bytes32)
    {
        return keccak256(prefix, _preimage);
    }
    
}
/* Submitting a payment, IDEAS

When the user submits a payment, they must provide a signed message, to ensure they can't game the pay system.

*/