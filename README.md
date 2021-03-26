# Neo4j<>Kafka Middleman
A service that concurrently listens for events on Kafka topics and safely executes queries on a Neo4j database. 

## Usage
### Configuration
Put your configuration in a file named config.yaml

```yaml
---
neo4j:
  uri: "localhost:8070"
  username: "neo4j"
  password: "neo4j"
kafka:
  client_id: "go-kafka-middleman"
  brokers:
    - "broker1.kafka-cluster.com"
    - "broker2.kafka-cluster.com"
topic-query:
  example-topic: "MATCH (n) RETURN n"
```

### Building 
You can build a very minimal docker image (<10MB) by running

```
docker build . -t neo4j-kafka-middleman
```

### Running
Run the docker container:
```
docker run neo4j-kafka-middleman --name middleman -d
```

