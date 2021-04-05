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
	"github.com/Azure/go-autorest/autorest/adal"
)

// validateAccKey : Validate storage account using account key
func (az *azurestorageFS) validateAccount() (err error) {
	az.serviceURL, err = getServiceURL(az)
	if err != nil {
		Logger.LogErr("Failed to create service URL")
		return err
	}
	
	if Config.IsAuthTypeMSI() {
		az.containerURL = az.serviceURL.NewContainerURL(*Config.BlobfuseConfig.StoreContainerName)
	} else {
		az.containerURL = azblob.NewContainerURL(*az.epURL, az.azPipeline)
	}
	Logger.LogErr("Container URL is %s", az.containerURL)
	marker := (azblob.Marker{})

	//var lst *azblob.ListBlobsHierarchySegmentResponse
	_, err = az.containerURL.ListBlobsHierarchySegment(context.Background(), marker, "/",
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

// getServiceURL : Get the service URL using the config
func getServiceURL(az *azurestorageFS) (serviceURL azblob.ServiceURL, err error) {
	if Config.IsAuthTypeAccKey() {
		az.azPipeline, err = getAccKeyPipeline()
		if err != nil {
			Logger.LogErr("Failed to create pipeline using storage key")
			return serviceURL, err
		}
	} else if Config.IsAuthTypeSAS() {
		az.azPipeline, err = getSASPipeline()
		if err != nil {
			Logger.LogErr("Failed to create pipeline using SAS")
			return serviceURL, err
		}
	} else if Config.IsAuthTypeMSI() {
		az.azPipeline, err = getMSIPipeline(*Config.BlobfuseConfig.ApplicationID, *Config.BlobfuseConfig.ResourceID, *Config.BlobfuseConfig.Resource)
		if err != nil {
			Logger.LogErr("Failed to create pipeline using MSI")
			return serviceURL, err
		}
	}

	endpoint := "blob"
	if *Config.BlobfuseConfig.StorageAccountADLS {
		endpoint = "dfs"
	}

	Logger.LogErr("Selected endpoint is %s", endpoint)
	if Config.IsAuthTypeAccKey() {
		az.epURL, err = url.Parse(fmt.Sprintf("https://%s.%s.core.windows.net/%s",
			*Config.BlobfuseConfig.StoreAccountName, endpoint,
			*Config.BlobfuseConfig.StoreContainerName))
	} else if Config.IsAuthTypeSAS() {
		az.epURL, err = url.Parse(fmt.Sprintf("https://%s.%s.core.windows.net/%s%s",
			*Config.BlobfuseConfig.StoreAccountName,
			endpoint,
			*Config.BlobfuseConfig.StoreContainerName,
			*Config.BlobfuseConfig.StoreAccountSAS))
	} else if Config.IsAuthTypeMSI() {
		az.epURL, err = url.Parse("https://" + *Config.BlobfuseConfig.StoreAccountName + ".blob.core.windows.net/")
	}

	if err != nil {
		Logger.LogErr("Failed to parse the URL (%s)", err.Error)
		return serviceURL, err
	}

	return azblob.NewServiceURL(*az.epURL, az.azPipeline), nil
}

func getAccKeyPipeline() (p pipeline.Pipeline, err error) {
	Logger.LogErr("Creating a key based pipeline")
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

func getSASPipeline() (p pipeline.Pipeline, err error) {
	Logger.LogErr("Creating a SAS based pipeline")
	c := azblob.NewAnonymousCredential()
	return azblob.NewPipeline(c, azblob.PipelineOptions{}), nil
}

func getMSIPipeline(applicationID, identityResourceID, resource string) (p pipeline.Pipeline, err error) {
	Logger.LogErr("Creating a MSI based pipeline")
	callbacks := func(token adal.Token) error { return nil }
	tokenCredentials, err := getOAuthToken(applicationID, identityResourceID, resource, callbacks)
	if err != nil {
		Logger.LogErr("Failed to get the Auth token (%s)", err.Error())
	}
	return azblob.NewPipeline(*tokenCredentials, azblob.PipelineOptions{}), nil
}

func fetchMSIToken(applicationID string, identityResourceID string, resource string, callbacks ...adal.TokenRefreshCallback) (*adal.ServicePrincipalToken, error) {
	msiEndpoint, _ := adal.GetMSIVMEndpoint()

	var spt *adal.ServicePrincipalToken
	var err error

	if applicationID == "" && identityResourceID == "" {
		spt, err = adal.NewServicePrincipalTokenFromMSI(msiEndpoint, resource, callbacks...)
	} else if applicationID != "" {
		spt, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint, resource, applicationID, callbacks...)
	} else if identityResourceID != "" {
		spt, err = adal.NewServicePrincipalTokenFromMSIWithIdentityResourceID(msiEndpoint, resource, identityResourceID, callbacks...)
	}

	if err != nil {
		Logger.LogErr("Failed to create token (%s)", err.Error())
		return nil, err
	}

	return spt, spt.Refresh()
}

func getOAuthToken(applicationID, identityResourceID, resource string, callbacks ...adal.TokenRefreshCallback) (*azblob.TokenCredential, error) {
	spt, err := fetchMSIToken(applicationID, identityResourceID, resource, callbacks...)
	if err != nil {
		Logger.LogErr("Failed to get MSI token (%s)", err.Error())
		return nil, err
	}

	// Refresh obtains a fresh token
	err = spt.Refresh()
	if err != nil {
		Logger.LogErr("Failed to Refresh the token (%s)", err.Error())
		return nil, err
	}

	tc := azblob.NewTokenCredential(spt.Token().AccessToken, func(tc azblob.TokenCredential) time.Duration {
		err := spt.Refresh()
		if err != nil {
			Logger.LogErr("Failed to Refresh the token (%s)", err.Error())
			return 0
		}

		// set the new token value
		tc.SetToken(spt.Token().AccessToken)
		Logger.LogErr("Token retreived is %s (%d)", spt.Token().AccessToken, spt.Token().Expires())
		
		// get the next token slightly before the current one expires
		return time.Until(spt.Token().Expires()) - 10*time.Second
	})

	return &tc, nil
}

var azPiplineOptions = azblob.PipelineOptions{
	// Set RetryOptions to control how HTTP request are retried when retryable failures occur
	Retry: azblob.RetryOptions{
		Policy:        azblob.RetryPolicyExponential, // Use exponential backoff as opposed to linear
		MaxTries:      3,                             // Try at most 3 times to perform the operation (set to 1 to disable retries)
		TryTimeout:    time.Second * 3600,            // Maximum time allowed for any single try
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
