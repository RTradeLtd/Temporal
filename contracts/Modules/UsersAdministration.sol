pragma solidity 0.4.24;

contract UsersAdministration {
    address public owner;
    address public admin;

    constructor() public {
        owner = msg.sender;
        admin = msg.sender;
    }

    modifier onlyOwner(address _sender) {
        require(_sender == owner);
        _;
    }

    modifier onlyAdmin(address _sender) {
        require(_sender == owner || _sender == admin);
        _;
    }

    function changeAdmin(
        address _newAdmin)
        public
        onlyOwner(_newAdmin)
        returns (bool)
    {
        admin = _newAdmin;
        return true;
    }

}