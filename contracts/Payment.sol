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

}