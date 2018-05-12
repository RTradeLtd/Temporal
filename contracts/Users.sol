pragma solidity 0.4.23;

import "./Modules/UsersAdministration.sol";
import "./Math/SafeMath.sol";

interface ERC20I {
    function transferFrom(address _owner, address _recipient, uint256 _amount) external returns (bool);
    function transfer(address _recipient, uint256 _amount) external returns (bool);
}

contract Users  is UsersAdministration {
    using SafeMath for uint256;

    ERC20I public rtI = ERC20I(address(0));

    struct UserStruct {
        address ethAddress;
        uint256 availableBalance;
        uint256 lockedBalance;
        bytes32[] uploadsArray;
        mapping (bytes32 => bool) uploads;
        bool registered;
        bool lockedForWithdrawal;
    }

    mapping (address => UserStruct) public users;

    modifier isLockedForWithdrawal(address _user) {
        require(users[_user].lockedForWithdrawal);
        _;
    }

    modifier notLockedForWithdrawal(address _user) {
        require(!users[_user].lockedForWithdrawal);
        _;
    }

    modifier isRegistered(address _addr) {
        require(users[_addr].registered);
        _;
    }

    modifier notRegistered(address _addr) {
        require(!users[_addr].registered);
        _;
    }

    function registerUploader(
        address _uploader)
        public
        notRegistered(_uploader)
        onlyAdmin(msg.sender)
        returns (bool)
    {
        users[_uploader].registered = true;
        return true;
    }

    function depositFunds(
        uint256 _amount)
        public
        isRegistered(msg.sender)
        returns (bool)
    {
        users[msg.sender].availableBalance = _amount;
        require(rtI.transferFrom(msg.sender, address(this), _amount));
        return true;
    }

    function userLockFunds(
        uint256 _amount)
        public
        isRegistered(msg.sender)
        returns (bool)
    {
        require(users[msg.sender].availableBalance >= _amount);
        users[msg.sender].availableBalance = users[msg.sender].availableBalance.sub(_amount);
        users[msg.sender].lockedBalance = users[msg.sender].lockedBalance.add(_amount);
        return true;
    }

    function paymentLockFunds(
        address _user,
        uint256 _amount)
        public
        notLockedForWithdrawal(_user)
        returns (bool)
    {
        // placeholder check, will need to be the payment channelcontract
        require(msg.sender == address(0));
        require(users[_user].availableBalance >= _amount);
        users[_user].availableBalance = users[_user].availableBalance.sub(_amount);
        users[_user].lockedBalance = users[_user].lockedBalance.add(_amount);
        return true;
    }

    function withdrawLockedFundsForPayment(
        address _user,
        uint256 _amount)
        public
        onlyAdmin(msg.sender)
        returns (bool)
    {
        require(users[_user].lockedBalance >= _amount);
        uint256 remainingAmount = users[_user].lockedBalance.sub(_amount);
        users[_user].lockedBalance = remainingAmount;
        require(rtI.transfer(msg.sender, _amount));
        return true;
    }

    function withdrawAvailableFunds()
        public
        isRegistered(msg.sender)
        returns (bool)
    {
        uint256 deposit = users[msg.sender].availableBalance;
        users[msg.sender].availableBalance = 0;
        require(rtI.transfer(msg.sender, deposit));
        return true;
    }

    function getLockedBalance(
        address _user)
        public
        view
        returns (uint256)
    {
        return users[_user].lockedBalance;
    }

    function getAvailableBalance(
        address _user)
        public
        view
        returns (uint256)
    {
        return users[_user].availableBalance;
    }
}