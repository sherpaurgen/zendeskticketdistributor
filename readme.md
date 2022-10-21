

This program will use zendesk ticket and search api to detect new tickets/Triage tickets to available agents: users with least number of open/pending ticket will be assigned with new tickets.
The UI of the api is a node/react app (see dockercompose file and source link) and uses data stored in postgres+api.
 
To create go executable
go build ./cmd/backapi
copy the binary "backapi" to ~/utils/

Step 1) 
```
mkdir ~/.triage 
mkdir ~/utils
```
Step 2) 
Copy 3 files  go binary backapi + zendb.sql and docker-compose.yml to ~/utils/

Step 3)
create file:-  
```
vim ~/.triage/agents.json
```

```
{  
  "subdomain":"mycompanysupport.zendesk.com/api/v2/",
    "agentlist": [
        "user1@example.com",
        "user2@example.com",
        "user3@example.com",
         "userN@example.com"]
}
```
Step 4)   Create zd.json with zendesk credentials
```
vim  ~/.triage/zd.json

{
    "user": "user1@example.com",
    "pass": "SecretpassX"
}
```
Check if files below 3 files are present 
```
cd ~/utils/utils % ls 
    backapi   docker-compose.yml   zendb.sql 
```    
Source here :https://github.com/sherpaurgen/dashboard

Step 5)

```
docker-compose up -d   # this will download container images and run 2 containers(postgres/node)
# check containers are up with command --> docker ps 
% docker ps    
CONTAINER ID   IMAGE              COMMAND                  CREATED             STATUS             PORTS                    NAMES
ad74e6903818   postgres:14.4      "docker-entrypoint.s…"   About an hour ago   Up About an hour   0.0.0.0:5432->5432/tcp   tablo_db_1
e42492f38446   bioniclts/ui:1.0   "docker-entrypoint.s…"   About an hour ago   Up About an hour   0.0.0.0:3000->3000/tcp   ui
```

Step 6) Run the executable
```
cd ~/utils/
./backapi
```
```
--------------------
docker-compose.yml
--------------------
version: "3.8"
services:
  app:
    # build:
    #   context: .
    ports:
      - 3000:3000
    image: bioniclts/ui:1.0
    container_name: ui
    command: npm start

  db:
    image: postgres:14.4
    restart: always
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=postgres
    ports:
      - '5432:5432'  
    volumes:
      - db-data:/var/lib/postgresql/data
      - ~/utils/zendb.sql:/docker-entrypoint-initdb.d/zendb.sql

volumes:
  db-data:
```


