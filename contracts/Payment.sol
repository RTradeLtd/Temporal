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

    // PaymentState will keep track of the state of a payment, nil means we havent seen th payment before
    enum PaymentState{ nil, pending, paid }
    // How payments can be made, RTC or eth
    enum PaymentMethod{ RTC, ETH }

    // This will contain all the information related to a particular payment
    struct PaymentStruct {
        address uploader;
        bytes32 paymentID;
        bytes32 hashedCID;
        uint256 retentionPeriodInMonths;
        uint256 paymentAmount;
        PaymentState state;
        PaymentMethod method;
    }

    // every payment ID and its associates contents
    mapping (bytes32 => PaymentStruct) public payments;
    // total number of payments for a user
    mapping (address => uint256) public numPayments;
    // mapping of total number of payments to payment IDs.
    // we start at 1 for the uint256 index
    // when payment number is 2, the index will be 2, and so on
    mapping (address => mapping (uint256 => bytes32)) public paymentIDs;

    event FilesContractSet(address _filesContractAddress);
    event UsersContractSet(address _usersContractAddress);

    // this event is fired when a payment is registered
    // we index the uploader addres so we can easily search evnent topics with their address as another filter
    event PaymentRegistered(
        address indexed _uploader,
        bytes32 _hashedCID,
        uint256 _retentionPeriodInMonths,
        uint256 _amount,
        bytes32 _paymentID
    );

    // this event is fired when a payment is received
    event PaymentReceived(address indexed _uploader, bytes32 _paymentID, uint256 _amount, PaymentMethod _method);
    event PaymentReceivedNoIndex(address _uploader, bytes32 _paymentID, uint256 _amount, PaymentMethod _method);
    // this event is fired when a payment is received

    // checks to see if the caller belogns to hte payment id
    modifier isUploader(bytes32 _paymentID, address _uploader) {
        require(payments[_paymentID].uploader == _uploader);
        _;
    }

    // checks to see whether or not hte payment has been paid
    modifier isPendingPayment(bytes32 _paymentID) {
        require(payments[_paymentID].state == PaymentState.pending);
        _;
    }

    // make sure they are paying enough
    modifier validPaymentAmount(bytes32 _paymentID, uint256 _payment) {
        require(payments[_paymentID].paymentAmount == _payment);
        _;
    }

    // make sure its the right payment method
    modifier validPaymentMethod(bytes32 _paymentID, PaymentMethod _method) {
        require(payments[_paymentID].method == _method);
        _;
    }

    // we will call this function to register their payment with the contract
    // this will later be changed to use signatures so we dont have to sbumit the payment registration, saving gas
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
        // construct the paymetn id
        bytes32 paymentID = keccak256(_uploader, _hashedCID, _retentionPeriodInMonths, numPayments[_uploader], now);
        // ensure this payment id hasn't been submitted before
        require(payments[paymentID].state == PaymentState.nil);
        // construct the payment struct and store it in the mapping
        payments[paymentID] = PaymentStruct({
            uploader: _uploader,
            paymentID: paymentID,
            hashedCID: _hashedCID,
            retentionPeriodInMonths:_retentionPeriodInMonths,
            paymentAmount: _amount,
            state: PaymentState.pending,
            method: PaymentMethod(_method)
        });
        // increase their number of pamyents
        numPayments[_uploader] = numPayments[_uploader].add(1);
        // map the payment number to the payment id for easy off-chain fetching
        paymentIDs[_uploader][numPayments[_uploader]] = paymentID;
        emit PaymentRegistered(_uploader, _hashedCID, _retentionPeriodInMonths, _amount, paymentID);
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
        emit PaymentReceived(msg.sender, _paymentID, _amount, PaymentMethod.ETH);
        emit PaymentReceivedNoIndex(msg.sender, _paymentID, _amount, PaymentMethod.ETH);
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
        emit PaymentReceived(msg.sender, _paymentID, _amount, PaymentMethod.RTC);
        emit PaymentReceivedNoIndex(msg.sender, _paymentID, _amount, PaymentMethod.RTC);
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