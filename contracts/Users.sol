pragma solidity 0.4.23;

import "./Modules/UsersAdministration.sol";
import "./Math/SafeMath.sol";

interface ERC20I {
    function transferFrom(address _owner, address _recipient, uint256 _amount) external returns (bool);
    function transfer(address _recipient, uint256 _amount) external returns (bool);
}


/**
    Current Limitations:
        Locked balance can't be unlocked, it can't only be spent (this will change)
*/

contract Users is UsersAdministration {

    using SafeMath for uint256;

    enum UserStateEnum { nil, registered, disabled }

    struct UserStruct {
        address uploaderAddress;
        uint256 availableEthBalance;
        uint256 availableRtcBalance;
        uint256 lockedEthBalance;
        uint256 lockedRtcBalance;
        UserStateEnum state;
    }

    mapping (address => UserStruct) public users;

    event UserRegistered(address indexed _uploader);
    event EthDeposited(address indexed _uploader, uint256 _amount);
    event EthWithdrawn(address indexed _uploader, uint256 _amount);
    event EthLocked(address indexed _uploader, uint256 _amount);

    modifier nonRegisteredUser(address _uploaderAddress) {
        require(users[_uploaderAddress].state == UserStateEnum.nil);
        _;
    }

    modifier isRegistered(address _uploaderAddress) {
        require(users[_uploaderAddress].state == UserStateEnum.registered);
        _;
    }

    modifier validAvailableEthBalance(address _uploaderAddress, uint256 _amount) {
        require(users[_uploaderAddress].availableEthBalance >= _amount);
        _;
    }
    
    function registerUser()
        public
        nonRegisteredUser(msg.sender)
        returns (bool)
    {
        users[msg.sender].state = UserStateEnum.registered;
        users[msg.sender].uploaderAddress = msg.sender;
        emit UserRegistered(msg.sender);
        return true;
    }


    function depositEther()
        public
        payable
        isRegistered(msg.sender)
        returns (bool)
    {
        require(msg.value > 0);
        users[msg.sender].availableEthBalance = users[msg.sender].availableEthBalance.add(msg.value);
        emit EthDeposited(msg.sender, msg.value);
        return true;
    }

    function withdrawFromAvailableBalance(
        uint256 _amount)
        public
        isRegistered(msg.sender)
        validAvailableEthBalance(msg.sender, _amount)
        returns (bool)
    {
        uint256 remaining = users[msg.sender].availableEthBalance.sub(_amount);
        users[msg.sender].availableEthBalance = remaining;
        emit EthWithdrawn(msg.sender, _amount);
        msg.sender.transfer(remaining);
        return true;
    }

    function lockPortionOfAvailableEthBalance(
        uint256 _amount)
        public
        isRegistered(msg.sender)
        validAvailableEthBalance(msg.sender, _amount)
        returns (bool)
    {
        uint256 remaining = users[msg.sender].availableEthBalance.sub(_amount);
        users[msg.sender].availableEthBalance = remaining;
        users[msg.sender].lockedEthBalance = users[msg.sender].lockedEthBalance.add(_amount);
        emit EthLocked(msg.sender, _amount);
        return true;
    }


}