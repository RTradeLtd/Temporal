pragma solidity 0.4.24;

/**
 Used by a user to pay for a file upload

 Current limitations:
    require admin to pre-register payment (this will be changed to use signatures and proofs to avoid having to preregister payments)
 */

import "./Modules/PaymentAdministration.sol";
import "./Modules/Utils.sol";
import "./Math/SafeMath.sol";

interface UserInterface {
    function paymentProcessorWithdrawEthForUploader(address _uploaderAddress, uint256 _amount, bytes32 _hashedCID) external returns (bool);

    function paymentProcessorWithdrawRtcForUploader(address _uploaderAddress, uint256 _amount, bytes32 _hashedCID) external returns (bool);
}

interface FilesInterface {
        function addUploaderForCid(address _uploader, bytes32 _hashedCID, uint256 _retentionPeriodInMonths) external returns (bool);
}
contract Payment is PaymentAdministration, Utils {

    using SafeMath for uint256;

    bytes private prefix = "\x19Ethereum Signed Message:\n32";

    UserInterface public uI;
    FilesInterface public fI;

    enum PaymentState{ nil, pending, paid }
    enum PaymentMethod{ RTC, ETH }

    struct PaymentStruct {
        address uploader;
        bytes32 paymentID;
        bytes32 hashedCID;
        uint256 retentionPeriodInMonths;
        uint256 paymentAmount;
        PaymentState state;
        PaymentMethod method;
    }

    mapping (bytes32 => PaymentStruct) public payments;
    mapping (address => uint256) public numPayments;
    mapping (address => mapping (uint256 => bytes32)) public paymentIDs;

    event FilesContractSet(address _filesContractAddress);
    event UsersContractSet(address _usersContractAddress);
    event PaymentRegistered(address indexed _uploader, bytes32 _hashedCID, uint256 _retentionPeriodInMonths, uint256 _amount, bytes32 _paymentID);
    event EthPaymentReceived(address indexed _uploader, bytes32 _paymentID, uint256 _amount);
    event RtcPaymentReceived(address indexed _uploader, bytes32 _paymentID, uint256 _amount);

    modifier isUploader(bytes32 _paymentID, address _uploader) {
        require(payments[_paymentID].uploader == _uploader);
        _;
    }

    modifier isPendingPayment(bytes32 _paymentID) {
        require(payments[_paymentID].state == PaymentState.pending);
        _;
    }

    modifier validPaymentAmount(bytes32 _paymentID, uint256 _payment) {
        require(payments[_paymentID].paymentAmount == _payment);
        _;
    }

    modifier validPaymentMethod(bytes32 _paymentID, PaymentMethod _method) {
        require(payments[_paymentID].method == _method);
        _;
    }

    function registerPayment(
        address _uploader,
        bytes32 _hashedCID,
        uint256 _retentionPeriodInMonths,
        uint256 _amount,
        uint8   _method) // 0 = RTC, 1 = ETH
        public
        onlyAdmin(msg.sender)
        greaterThanZeroU(_amount)
        greaterThanZeroU(_retentionPeriodInMonths)
        nonZeroAddress(_uploader)
        returns (bool)
    {
        if (_method > 1 || _method < 0) { // _method can only be 0 or 1
            revert();
        }
        bytes32 paymentID = keccak256(_uploader, _hashedCID, _retentionPeriodInMonths, numPayments[_uploader]);
        require(payments[paymentID].state == PaymentState.nil);
        payments[paymentID] = PaymentStruct({
            uploader: _uploader,
            paymentID: paymentID,
            hashedCID: _hashedCID,
            retentionPeriodInMonths:_retentionPeriodInMonths,
            paymentAmount: _amount,
            state: PaymentState.pending,
            method: PaymentMethod(_method)
        });
        numPayments[_uploader] = numPayments[_uploader].add(1);
        paymentIDs[_uploader][numPayments[_uploader]] = paymentID;
        emit PaymentRegistered(msg.sender, _hashedCID, _retentionPeriodInMonths, _amount, paymentID);
        return true;
    }

    function payEthForPaymentID(
        bytes32 _paymentID,
        uint256 _amount)
        public
        isPendingPayment(_paymentID)
        isUploader(_paymentID, msg.sender)
        validPaymentAmount(_paymentID, _amount)
        validPaymentMethod(_paymentID, PaymentMethod.ETH)
        returns (bool)
    {
        payments[_paymentID].state = PaymentState.paid;
        emit EthPaymentReceived(msg.sender, _paymentID, _amount);
        require(fI.addUploaderForCid(msg.sender, payments[_paymentID].hashedCID, payments[_paymentID].retentionPeriodInMonths));
        require(uI.paymentProcessorWithdrawEthForUploader(msg.sender, _amount, payments[_paymentID].hashedCID));
        return true;
    }

    function payRtcForPaymentID(
        bytes32 _paymentID,
        uint256 _amount)
        public
        isPendingPayment(_paymentID)
        isUploader(_paymentID, msg.sender)
        validPaymentAmount(_paymentID, _amount)
        validPaymentMethod(_paymentID, PaymentMethod.RTC)
        returns (bool)
    {
        payments[_paymentID].state = PaymentState.paid;
        emit RtcPaymentReceived(msg.sender, _paymentID, _amount);
        require(fI.addUploaderForCid(msg.sender, payments[_paymentID].hashedCID, payments[_paymentID].retentionPeriodInMonths));
        require(uI.paymentProcessorWithdrawRtcForUploader(msg.sender, _amount, payments[_paymentID].hashedCID));
        return true;
    }

    function setFilesInterface(
        address _filesContractAddress)
        public
        onlyAdmin(msg.sender)
        nonZeroAddress(_filesContractAddress)
        returns (bool)
    {
        fI = FilesInterface(_filesContractAddress);
        emit FilesContractSet(_filesContractAddress);
        return true;
    }

    function setUsersInterface(
        address _usersContractAddress)
        public
        onlyAdmin(msg.sender)
        returns (bool)
    {
        uI = UserInterface(_usersContractAddress);
        emit UsersContractSet(_usersContractAddress);
        return true;
    }
}