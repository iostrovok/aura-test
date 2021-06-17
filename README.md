# aura

### Install and run server

You should have installed GO version >= 1.13.

	git clone https://github.com/iostrovok/aura-test
    cd aura-test 
    make run

It starts server on localhost:8080.

### Run test scripts

Open new console window and go to aura-test folder.

    go run ./console/actions/main.go

or

    go run ./console/simple_load/simple_load.go

### Protocol

#### Create new session.

    Method "POST"
    URL "/sessions"
    Parameter "TTL" optional, positive integer, ttl <= 30

#### List of all sessions.

    Method "GET"
    URL "/sessions"

#### Destroy the session with session id "id".

    Method "DELETE"
    URL "/sessions/{id}"

#### Extend the session with session id "id".

    Method "PUT"
    URL "/sessions/{id}" (default TTL 30 sec))
    URL "/sessions/{id}/{ttl}" (ttl is positive integer, 0 < ttl <= 300)

### Examples

#### Create new session with default TTL (30 sec)

    curl -X POST  http://localhost:8080/sessions

#### Create new session with TTL = 5

    curl -X POST -d 'TTL=5' http://localhost:8080/sessions

#### List of all sessions

    curl -XGET 'http://localhost:8080/sessions'

#### Expend the session TTL with default TTL (30 sec)

    curl -XPUT 'http://localhost:8080/sessions/<id>'

#### Extend the session TTL with TTL = 100

    curl -XPUT 'http://localhost:8080/sessions/<id>/100'

### Destroy the session

    curl -XDELETE 'http://localhost:8080/sessions/<id>'

#### See ./actions folder and tests for more examples. 