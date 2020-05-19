# Service Tester
Service to test icmp, tcp, and udp traffic across multiple containers world wide. Primarily used for testing the StackPath Edge Compute platform.


![Dashboard](https://raw.githubusercontent.com/brentahughes/service_tester/master/screenshots/dashboard.png)
![Check Details](https://raw.githubusercontent.com/brentahughes/service_tester/master/screenshots/check_details.png)

## How It Works
Every few seconds the service will do a dns query to a discovery dns to find all other containers that are a part of the test as defined by the envvar DISCOVERY_NAME. For each container it finds it will begin a periodic health, icmp, tcp, and udp check on every host it's found. This creates of every container talking to every other container constantly.

Every check it does is stored in the database with the status and response time.


## Startup

```
 docker run --rm -p 80:80 -p 5500:5500 -p 5500:5500/udp -e "DISCOVERY_NAME=dns.address.for.container.list" service_tester
```

| Ports | Description |
| ----- | ----------- |
| 80    | web interface |
| 5500  | tcp/udp service |

DISCOVERY_NAME is should be an A record that returns a list of IPs. This is NOT a SRV record.

## Development

### Backend API
Must be run as root or with super user privileges for ICMP tests to work.
`DISCOVERY_NAME="dns.address.for.container.list" go run main.go`

This will also startup the frontend but it will the the production build of the frontend and should not be used for development

### Frontend
The frontend is a ReactJS app that talks to the backend api. In production it is served by the same go server that the backend uses and is served as a static html file

`cd frontend && npm start`

This will start a development server of the frontend on `localhost:3000`. It requires the backend to be running for api requests to success.

You can build the production frontend with `npm run --prefix frontend build` which will create the static html file that is served by the go service.
