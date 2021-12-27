./fnua --config=./flow.yml subscribe time.flows.unix.ar --nameserver 172.17.0.2 -d --agent fnaa-emiliano
./fnua --config=./flow.yml create flow time.flows.unix.ar --nameserver 172.17.0.2 -d
./fnua --config=./flow.yml describe flow time.flows.unix.ar --nameserver 172.17.0.2 -d
