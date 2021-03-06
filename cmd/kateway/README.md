# kateway

A fully-managed real-time secure and reliable RESTful Cloud Pub/Sub streaming message/job service.

    _/    _/              _/
       _/  _/      _/_/_/  _/_/_/_/    _/_/    _/      _/      _/    _/_/_/  _/    _/
      _/_/      _/    _/    _/      _/_/_/_/  _/      _/      _/  _/    _/  _/    _/
     _/  _/    _/    _/    _/      _/          _/  _/  _/  _/    _/    _/  _/    _/
    _/    _/    _/_/_/      _/_/    _/_/_/      _/      _/        _/_/_/    _/_/_/
                                                                               _/

### Alternatives

- Google Cloud Pub/Sub
- Amazon kenesis
- Azure EventHub
- IBM Bluemix Message Hub
- misc
  - pubnub
  - pusher
  - firebase
  - parse
  - pubsubhubbub

### Features

- REST API
  - http/https/websocket/http2 interface for Pub/Sub
- Support both FIFO and Priority queue
- Systemic Quality Requirements
  - Performance & Throughput
    - > 100K msg/sec delivery on a single host without batch
  - Scalability
    - scales to 1M msg/sec
    - elastic scales
  - Latency
    - < 1s delivery
  - Availability
    - Graceful shutdown without downtime
  - Graceful Degrade
    - throttle
    - circuit breaker
  - Rich monitoring and alarm 
- Fully-managed
  - Discovery
  - Create versioned topics, subscribe to topics
  - Dedicated real-time metrics and fully-functional dashboard 
  - Easy trouble shooting
  - Visualize message flow
  - [ ] Managed integration service via Webhooks
- Communication can be 
  - one-to-many (fan-out)
  - many-to-one (fan-in)
  - many-to-many
- Replicated storage and guaranteed at-least-once message delivery
- Flexible delivery options
  - Both push- and pull-style subscriptions supported
- Enables sophisticated streaming data processing
  - because one app may emit kateway stream data into another kateway stream
- [ ] Quotas and rate limit, QoS
  - Flow control: Dynamic rate limiting
- [ ] Encryption of all message data on the wire


### Common scenarios

- Balancing workloads in network clusters
- Implementing asynchronous workflows
- Distributing event notifications
- Refreshing distributed caches
- Logging to multiple systems
- Data streaming from various processes or devices
- Reliability improvement

### Design philosophy

kateway is designed as an open system that can easily integrate with other system.

It is designed to be programmer friendly.

### Architecture

           +----+      +-----------------+          
           | DB |------|   manager UI    |
           +----+      +-----------------+                                                  
                               |                                                           
                               ^ register [application|topic|version|subscription]                       
                               |                                                          
                       +-----------------+                                                 
                       |  Application    |                                                
                       +-----------------+                                               
                               |                                                        
                               V                                                       
            PUB                |               SUB                                    
            +-------------------------------------+                                  
            |                                     |                                         
       HTTP |                                HTTP | keep-alive 
       POST |                                 GET | session sticky                        
            |                                     |                                      
        +------------+                      +------------+                 application 
     ---| PubEndpoint|----------------------| SubEndpoint|---------------------------- 
        +------------+           |          +------------+                     kateway
        | stateless  |        Plugins       | stateful   |                           
        +------------+  +----------------+  +------------+                          
        | quota      |  | Authentication |  | quota      |      
        +------------+  +----------------+  +------------+     
        | metrics    |  | Authorization  |  | metrics    |    
        +------------+  +----------------+  +------------+   
        | guard      |  | ......         |  | guard      |  
        +------------+  +----------------+  +------------+                      
        | registry   |                      | registry   |  
        +------------+                      +------------+                      
        | meta       |                      | meta       |  
        +------------+                      +------------+                      
        | manager    |                      | manager    |  
        +------------+                      +------------+                      
            |                                     |    
            |    +----------------------+         |  
            |----| ZK or other ensemble |---------| 
            |    +----------------------+         |
            |                                     |    
            | Append                              | Fetch
            |                                     |                     
            |       +------------------+          |     
            |       |      Store       |          |    
            +-------+------------------+----------+   
                    |  kafka or else   |
           +----+   +------------------+        +---------------+
           | gk |---|     monitor      |--------| elastic scale |
           +----+   +------------------+        +---------------+


### APIs

#### Pub

    POST    /v1/msgs/:topic/:ver
    POST /v1/ws/msgs/:topic/:ver

    POST    /v1/jobs/:topic/:ver
    POST /v1/ws/jobs/:topic/:ver
    DELETE  /v1/jobs/:topic/:ver

#### Sub

    GET    /v1/msgs/:appid/:topic/:ver
    GET /v1/ws/msgs/:appid/:topic/:ver

    POST   /v1/shadow/:appid/:topic/:ver/:group
    DELETE /v1/groups/:appid/:topic/:ver/:group

    GET /v1/subd/:topic/:ver
    GET /v1/status/:appid/:topic/:ver

#### Management

    GET    /alive
    GET    /v1/status
    GET    /v1/clusters
    GET    /v1/clients
    GET    /v1/partitions/:cluster/:appid/:topic/:ver
    POST   /v1/topics/:cluster/:appid/:topic/:ver
    DELETE /v1/counter/:name

### FAQ

- why named kateway?

  Admittedly, it is not a good name. Just short for kafka gateway

- how to batch messages in Pub?

  It is http client's job to put the variant length data into json array

- how to consume multiple messages in Sub?

  kateway uses chunked transfer encoding

- http header size limit?

  4KB

- what is limit of a pub message in size?

  1 ~ 256KB

- if sub with no arriving message, how long do client get http 204?

  30s

### Migration

    pub write both kafka and kateway
    pub write EOF to kafka while writing START to kateway atomically (for multiple partitions?)
        from then only pub write only to kateway
    sub(kafka) handles messages till encounter EOF
    sub(kateway) not handles message till encounter START

### TODO

- [ ] sub: what if broker moved
- [ ] pub in batch
- [X] change all handler declaration to xxxServer instead of Gateway
- [X] cf.ChannelBufferSize=20
- [X] topic name obfuscation
- [ ] sub with delayed ack
  - StatusNotModified
  - what if rebalanced, and ack buffered p/o
- [ ] test max body/header size limit
- [ ] refactor api pkg
- [ ] bugs
  - [ ] sit 36.topic-owl-biz.v1 has 20 partitions, cg only consumes part of it
  - [ ] kateway/clients.go need work for w-haproxy
- [ ] verify
- [ ] max request for a single persitent connection
  - make load balancer distribute req more evenly
- [ ] pub/sub a disabled topic, discard?
- [ ] features confirm
  - delayed job
  - bury
  - msg tag
- [ ] fetchShadowQueueRecords enable
- [ ] why make sub with 15s sleep fails
- [ ] Plugins
  - authentication and authorization
  - transform
  - hooks
  - other stuff related to message-oriented middleware
- [ ] schema registry with Avro
- [ ] check hack pkg
- [ ] https://github.com/allinurl/goaccess
- [ ] https, outer ip must https
- [ ] Update to glibc 2.20 or higher
- [ ] compress
  - gzip sub response
  - Pub HTTP POST Request compress
  - compression in kafka
