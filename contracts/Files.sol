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

    struct UserUploads {
        string ipfsHash;
        uint256 releaseDate;
    }

    mapping (bytes32 => FileUpload) public uploads;
    mapping (address => UserUploads[]) public userUploads;
    mapping (address => mapping (string => uint256)) private userUploadIndexes;
}

/**

    I've spec'd out the file upload struct, i'd like you to write functions to update it
 */