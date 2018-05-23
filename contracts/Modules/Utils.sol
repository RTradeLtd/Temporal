pragma solidity 0.4.24;

contract Utils {

    modifier greaterThanZeroU(uint256 _value) {
        require(_value > 0);
        _;
    }

    modifier nonZeroAddress(address _addr) {
        require(_addr != address(0));
        _;
    }
}