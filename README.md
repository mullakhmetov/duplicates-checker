# duplicates-checker

* [Specification](#spec)
* [Usage](#usage)
  * [Generate and load dataset](#usage-generate)
  * [Start REST](#usage-rest)


<a name="spec"></a>
## Specification

Implement REST service.

It will accept two user ids and respond with boolean value, which will be true if theese users are duplicates.

Duplicates are pairs of `user_id`s for which there are at least two matching ip adresses in the access log.
`user_id` is considered a duplicate of itself.
Each user can have multiple access records, and it is perfectly ok if there are many accesses from single ip address.
There are no unique constraints in access log at all.
Number of requests per user varies greatly from 1 to million or even more.
Number of different IPs user uses are on other hand rather small - most users have 1 or 2 distinct IPs.
Access log can be generated randomly in database or in plaintext file - it's up to you to decide.

Log format is rougly like this: `create table conn_log (user_id bigint, ip_addr varchar(15))`
IPs are in a regular IPv4 format (4 octets in decimal delimited by dots).
There should be no less than 10 millions of records in access log.

Service response time **should not exceed 5ms**

You should write your code as if it is a production service.

Example:

There are such records in conn_log:

```
1, 127.0.0.1
2, 127.0.0.1
1, 127.0.0.2
2, 127.0.0.2
2, 127.0.0.3
3, 127.0.0.3
3, 127.0.0.1
4, 127.0.0.1
```

Get request: http://localhost:8080/1/2
Response:
```json
{ "dupes": true }
```

Get request: http://localhost:8080/1/3
Response:

```json
{ "dupes": false }
```

Get request: http://localhost:8080/2/1
Response:

```json
{ "dupes": true }
```

Get request: http://localhost:8080/2/3
Response:

```json
{ "dupes": true }
```

Get request: http://localhost:8080/3/2
Response:

```json
{ "dupes": true }
```

Get request: http://localhost:8080/1/4
Response:

```json
{ "dupes": false }
```

Get request: http://localhost:8080/3/1
Response:

```json
{ "dupes": false}
```

Get request: http://localhost:8080/1/1
Response:

```json
{ "dupes": true}
```

<a name="usage"></a>
## Usage

#### build
```bash
make
```

#### help
```
./duplicates-checker --help
```

<a name="usage-generate"></a>
### generate and load dataset

#### help
`./duplicates-checker import --help`
#### generate dataset for debug
`./duplicates-checker import --dbg`


<a name="usage-rest"></a>
### start REST

#### help
`./duplicates-checker server --help`
#### run in debug mod on 8080 port
`./duplicates-checker server --port=8080 --dbg`

### make request
`curl http://localhost:8080/1/2/`
