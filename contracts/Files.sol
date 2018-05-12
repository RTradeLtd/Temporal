pragma solidity 0.4.23;

/**
    This contract is used to act as a central, immmutable record of the files stored in our system, allowing independent audits of our data storage
*/

contract FileRepository {


    struct FileUpload {
        string ipfsHash; // the actual content hash of the file, if a directory it will be the parent directory hash
        uint256 releaseDate; // refers to the time at which we will no longer pin the file in our syste
        address[] uploadersArray;
        mapping (address => bool) uploaders;
    }

    mapping (bytes32 => FileUpload) public uploads;
}