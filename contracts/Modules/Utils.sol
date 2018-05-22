pragma solidity 0.4.23;

contract Utils {

    modifier greaterThanZeroU(uint256 _value) {
        require(_value > 0);
        _;
    }
}