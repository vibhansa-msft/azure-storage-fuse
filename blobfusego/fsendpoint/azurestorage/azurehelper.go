package azurestorage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	Config "github.com/blobfusego/global"
	Logger "github.com/blobfusego/global/logger"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

// getServiceURL : Get the service URL using the config
func getServiceURL() (serviceURL azblob.ServiceURL, err error) {
	var pipeline pipeline.Pipeline

	if Config.IsAuthTypeAccKey() {
		pipeline, err = getAccKeyPipeline()
		if err != nil {
			Logger.LogErr("Failed to create pipeline using storage key")
			return serviceURL, err
		}
	}

	endpoint := "blob"
	if *Config.BlobfuseConfig.StorageAccountADLS {
		endpoint = "dfs"
	}

	Logger.LogDebug("Selected endpoint is %s", endpoint)

	url, err := url.Parse(fmt.Sprintf("https://%s.%s.core.windows.net",
		*Config.BlobfuseConfig.StoreAccountName, endpoint))

	if err != nil {
		Logger.LogErr("Failed to parse the URL (%s)", err.Error)
		return serviceURL, err
	}

	return azblob.NewServiceURL(*url, pipeline), nil
}

func getAccKeyPipeline() (p pipeline.Pipeline, err error) {

	credential, err := azblob.NewSharedKeyCredential(
		*Config.BlobfuseConfig.StoreAccountName,
		*Config.BlobfuseConfig.StoreAccountKey)

	if credential == nil || err != nil {
		Logger.LogDebug("Failed to create credential %s", err.Error())
		return p, err
	}

	// Create pipeline to intialize factories in sdk for retry logic
	return azblob.NewPipeline(credential, azPiplineOptions), nil
}

// validateAccKey : Validate storage account using account key
func (az *azurestorageFS) validateAccount() error {
	serviceURL, err := getServiceURL()
	if err != nil {
		Logger.LogErr("Failed to create service URL")
		return err
	}
	az.serviceURL = serviceURL
	az.containerURL = az.serviceURL.NewContainerURL(*Config.BlobfuseConfig.StoreContainerName)

	az.ctx = context.Background()
	marker := (azblob.Marker{})

	//var lst *azblob.ListBlobsHierarchySegmentResponse
	_, err = az.containerURL.ListBlobsHierarchySegment(az.ctx, marker, "/",
		azblob.ListBlobsSegmentOptions{MaxResults: 2})

	if err != nil {
		Logger.LogErr("Failed to validate account with given auth %s", err.Error)
		return err
	}

	/*
		for _, blob := range lst.Segment.BlobItems {
			Logger.LogDebug("GOT %s", blob.Name)
		}
	*/
	return nil
}

var azPiplineOptions = azblob.PipelineOptions{
	// Set RetryOptions to control how HTTP request are retried when retryable failures occur
	Retry: azblob.RetryOptions{
		Policy:        azblob.RetryPolicyExponential, // Use exponential backoff as opposed to linear
		MaxTries:      3,                             // Try at most 3 times to perform the operation (set to 1 to disable retries)
		TryTimeout:    time.Second * 3,               // Maximum time allowed for any single try
		RetryDelay:    time.Second * 1,               // Backoff amount for each retry (exponential or linear)
		MaxRetryDelay: time.Second * 3,               // Max delay between retries
	},

	/*
		    // Set RequestLogOptions to control how each HTTP request & its response is logged
		    RequestLog: RequestLogOptions{
		        LogWarningIfTryOverThreshold: time.Millisecond * 200, // A successful response taking more than this time to arrive is logged as a warning
		    },

		    // Set LogOptions to control what & where all pipeline log events go
		    Log: pipeline.LogOptions{
		        Log: func(s pipeline.LogLevel, m string) { // This func is called to log each event
		            // This method is not called for filtered-out severities.
		            logger.Output(2, m) // This example uses Go's standard logger
		        },
		        ShouldLog: func(level pipeline.LogLevel) bool {
		            return level <= pipeline.LogWarning // Log all events from warning to more severe
		        },
		    },

		    // Set HTTPSender to override the default HTTP Sender that sends the request over the network
		    HTTPSender: pipeline.FactoryFunc(func(next pipeline.Policy, po *pipeline.PolicyOptions) pipeline.PolicyFunc {
		        return func(ctx context.Context, request pipeline.Request) (pipeline.Response, error) {
		            // Implement the HTTP client that will override the default sender.
		            // For example, below HTTP client uses a transport that is different from http.DefaultTransport
		            client := http.Client{
		                Transport: &http.Transport{
		                    Proxy: nil,
		                    DialContext: (&net.Dialer{
		                        Timeout:   30 * time.Second,
		                        KeepAlive: 30 * time.Second,
		                        DualStack: true,
		                    }).DialContext,
		                    MaxIdleConns:          100,
		                    IdleConnTimeout:       180 * time.Second,
		                    TLSHandshakeTimeout:   10 * time.Second,
		                    ExpectContinueTimeout: 1 * time.Second,
		                },
		            }

		            // Send the request over the network
		            resp, err := client.Do(request.WithContext(ctx))

		            return pipeline.NewHTTPResponse(resp), err
		        }
			}),
	*/
}
