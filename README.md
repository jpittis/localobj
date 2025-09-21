Provides a testing library for spawning and managing local object store instances. It
supports pluggable S3-compatible stores with MinIO as the default.

## Usage

```
func TestExample(t *testing.T) {
	ctx := t.Context()

	store, err := StartDefaultStore(ctx)
	require.NoError(t, err)
	defer store.GracefulShutdown(ctx)

	client, err := store.NewClient()
	require.NoError(t, err)

	bucket := "test-bucket"
	key := "test-object"
	content := "test content"

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	require.NoError(t, err)

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
	})
	require.NoError(t, err)

	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	require.NoError(t, err)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, content, string(data))
}
```
