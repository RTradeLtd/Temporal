pragma solidity 0.4.23;

interface ERC20I {
    function transferFrom(address _owner, address _recipient, uint256 _amount) external returns (bool);
}

contract Users {

    ERC20I public rtI = ERC20I(address(0));

    struct UserStruct {
        address ethAddress;
        uint256 availableBalance;
        uint256 lockedBalance;
        bytes32[] uploadsArray;
        mapping (bytes32 => bool) uploads;
        bool registered;
    }

    mapping (address => UserStruct) public users;

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
        return true;
    }
}