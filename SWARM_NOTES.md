# Swarm

# Links

https://swarm-guide.readthedocs.io/en/latest/runninganode.html#connecting-to-the-swarm-testnet

https://swarm-guide.readthedocs.io/en/latest/runninganode.html#running-a-private-swarm

https://swarm-guide.readthedocs.io/en/latest/architecture.html

# Info

Swarm uses a HTTP API


# Examples

## Upload

curl -H "Content-Type: text/plain" --data-binary "some-data" http://localhost:8500/bzz:/

That will return a hex string similar to "027e57bcbae76c4b6a1c5ce589be41232498f1af86e1b1a2fc2bdffd740e9b39" which is the swarm content hash

## Download

curl http://localhost:8500/bzz:/027e57bcbae76c4b6a1c5ce589be41232498f1af86e1b1a2fc2bdffd740e9b39/

## Tar Stream Upload

( mkdir dir1 dir2; echo "some-data" | tee dir1/file.txt | tee dir2/file.txt; )

tar c dir1/file.txt dir2/file.txt | curl -H "Content-Type: application/x-tar" --data-binary @- http://localhost:8500/bzz:/
> 1e0e21894d731271e50ea2cecf60801fdc8d0b23ae33b9e808e5789346e3355e

## Multi Download Get Request

curl -s -H "Accept: application/x-tar" http://localhost:8500/bzz:/ccef599d1a13bed9989e424011aed2c023fce25917864cd7de38a761567410b8/ | tar t
> dir1/file.txt
  dir2/file.txt
  dir3/file.txt

  ## Multipart Form Upload

  curl -F 'dir1/file.txt=some-data;type=text/plain' -F 'dir2/file.txt=some-data;type=text/plain' http://localhost:8500/bzz:/
> 9557bc9bb38d60368f5f07aae289337fcc23b4a03b12bb40a0e3e0689f76c177

## Adding to an existing manifest

curl -F 'dir3/file.txt=some-other-data;type=text/plain' http://localhost:8500/bzz:/9557bc9bb38d60368f5f07aae289337fcc23b4a03b12bb40a0e3e0689f76c177
> ccef599d1a13bed9989e424011aed2c023fce25917864cd7de38a761567410b8

## Listing Files

curl -s http://localhost:8500/bzz-list:/ccef599d1a13bed9989e424011aed2c023fce25917864cd7de38a761567410b8/ | jq .