# fizzbuzz-server-go
Implementation of the fizzbuzz game in Go - https://en.wikipedia.org/wiki/Fizz_buzz

## Run locally

First thing to do is clone the repository of your choice
```
git clone https://github.com/julien-beguier/fizzbuzz-server-go.git
```
Change directory and launch:
```
docker-compose up --build
```

That's it!

***During the compose step, the http-server will attempt to connect to the mysql container but since it's initialization is slow it can take some time. It will throw errors because it can't immediately connect but this is normal. This only happens on the first launch.***

You can now request either:
```
localhost:8080/list
```
or
```
localhost:8080/statistics
```

## Available routes:

### `GET` /list

```
/list?limit=100&int1=3&int2=5&str1=toto&str2=titi
```

* Returns a list of strings with numbers based on the fizzbuzz algorithm.

From 1 to limit, where: all multiples of int1 are replaced by str1, all
multiples of int2 are replaced by str2, all multiples of int1 and int2 are
replaced by str1str2.

Query parameters

* limit: limit is the max number of the list. Range from 1 to Integer.MAX_VALUE.
* int1: int1 is the first number to check multiples for. Range from 1 to Integer.MAX_VALUE.
* int2: int2 is the first number to check multiples for. Range from 1 to Integer.MAX_VALUE.
* str1: str1 is the String that replace the multiples of int1
* str2: str2 is the String that replace the multiples of int2


### `GET` /statistics

```
/statistics
```

* Expects no parameter
* Returns the parameters corresponding to the most used request, as well as the
number of hits for this request

## Requirements

docker & docker-compose