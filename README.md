# learn_elasticsearch

### run
```bash
go run main.go
```
### Explanation

My first hands on using elastic search, this is golang implementation of 3 articles below: 
1. [Prefix Search](https://blog.mimacom.com/autocomplete-elasticsearch-part1/)
2. [N-Gram Search](https://blog.mimacom.com/autocomplete-elasticsearch-part2/)
3. [Completion Suggester](https://blog.mimacom.com/autocomplete-elasticsearch-part3/)

This uses v7.x.x

### Flow

1. Create a client by connecting to ES server
2. Show info
3. Check if index already exist, if yes, delete it
4. Create the index using standard analyzers
5. Search using prefix & Fuzzy
6. Search using match on prefix
7. Create another index using n-gram analyzers
8. 