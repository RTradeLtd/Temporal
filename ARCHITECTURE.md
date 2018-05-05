# Architecture


## IPFS

We will operate an IPFS cluster initially consisting of three nodes, two at HQ one off-site. The two nodes at HQ will exist on physically seperated devices to ensure protection against hardware failures. The third node will exist off-site, in case of network or environmental failures data availability will still exist for our customers while we work on bringing back the failed clusters.

Everytime content is uploaded to our system, we add the hashes to our database to create a easily indexable, and backupable listing of file in our system. Periodically we will parse through this database to ensure the files in our clusters are still there. If we notice any discrepancies, missing files will be re-inserted into the cluster

### Data Backup (expand on)

Once a day we will incremental backups of our data, encrypting them with our public PGP key. Once a week we will do  afull backup of our data, encrypting with a public PGP key.

### Nodes

Nodes will be running the golang reference implementation of the IPFS spec, and the golang implementation and will run on Ubuntu 16.04.3LTS.

### Security Considerations

### How It Works

When a file upload, or pin request is made to the system, we check to ensure that the file was paid for (we do this by looking up a sha256 hash of the file) in our smart contract, along with the ethereum address of the uploader. If it is a pin request, then we look up the pinned hash and the uploaders ethereum address and make sure it was paid for. If everything checks out the content is added to our system.

Hold up a minute, files, hashes? I thought IPFS was a content addressed data storage system. It is! However we make use of IPFS for two different purposes. One is to ensure that content people want pinned and accessible, but may not be able to pin regularly themselves is still accessibl; In this model you could be the primary storer of your data, and simply want us to act as a secondary host, backup of your content or we could be the primary provider of the content hash, and you the secondary. The possibilities are endless! The second way we make use of IPFS, is for the data storage, and replication of files that people upload to our storage farms. We decided to  go with IPFS due to it's immense capabilities, promising future, and ease of integration into our infrastructure due to a majority of the protocol being built with golang. 


To prevent abuse of the pricing system, even if a file or hash is already pinned on the system, a subsequent pin request from a different user will incur data charges according to how long that file or hash is to be pinned in our system, since that user is also requesting data persistence. In terms of files remaining in our system, the longest pin request is what we follow. 

In order to prevent issues with pinning large files to clusters, and timeouts we first pin to an ipfs node locally, followed by adding that pin to the cluster pinset.

### Pricing

Currently we only issue charges in one month intervals. To prevent people from uploading a file to our system and not paying the monthly fee, when a file is added to our system the RTC required to pay will be locked into a smart contract. We will only withdraw from this once a month. Whatever RTC is remaining after the time expires, you are free to withdraw that. If we forget to withdraw our fees, then that's on us not on you!

Should you wish to pay for these services without RTC and use other cryptos, that is also possible but will incur a 5% markup

### Scalability

This is an on-going considering for us, and we are always analyzing how to ensure we can scale as much as possible