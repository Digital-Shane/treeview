# s3 extension

This module provides a TreeView constructor for AWS S3 buckets.

## Example

```go
	// Create the tree with default options
	tree, err := s3.NewTreeFromS3(context.Background(),  "s3://my-bucket-name",  "default")
	if err != nil {
		log.Fatal(err)
	}
	// Render the tree to a string & print it
	output, _ := tree.Render(context.Background())
	fmt.Println(output)
```

## Contributing

The tests for this module require [localstack](https://www.localstack.cloud/).