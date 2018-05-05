# Architecture


Our main cluster is a set of two nodes with local storage. They are accessed through a load balancer that distributes requests to each of the nodes round robin.

We then have a remote node, serving as archival 