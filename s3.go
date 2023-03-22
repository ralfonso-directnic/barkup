package barkup

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

// S3 is a `Storer` interface that puts an ExportResult to the specified S3 bucket. Don't use your main AWS keys for this!! Create read-only keys using IAM
type S3 struct {
	// Available regions:
	// * us-east-1
	// * us-west-1
	// * us-west-2
	// * eu-west-1
	// * ap-southeast-1
	// * ap-southeast-2
	// * ap-northeast-1
	// * sa-east-1
	Region string
	// Name of the bucjet
	Bucket string
	// AWS S3 access key
	AccessKey string
	// AWS S3 secret
	ClientSecret string
	// Use AWS CLI to upload to S3
	UseAwsCli bool
}

//add a custom aws region for services that have a s3 like api
/*

type Region struct {
	Name                 string // the canonical name of this region.
	EC2Endpoint          string
	S3Endpoint           string
	S3BucketEndpoint     string // Not needed by AWS S3. Use ${bucket} for bucket name.
	S3LocationConstraint bool   // true if this region requires a LocationConstraint declaration.
	S3LowercaseBucket    bool   // true if the region requires bucket names to be lower case.
	SDBEndpoint          string
	SNSEndpoint          string
	SQSEndpoint          string
	IAMEndpoint          string
	Sign                 Signer // Method which will be used to sign requests.
}

	Region{
	"us-east-1",
	"https://ec2.us-east-1.amazonaws.com",
	"https://s3.amazonaws.com",
	"",
	false,
	false,
	"https://sdb.amazonaws.com",
	"https://sns.us-east-1.amazonaws.com",
	"https://sqs.us-east-1.amazonaws.com",
	"https://iam.amazonaws.com",
	SignV2,
}
	

*/

func (x *S3) CustomRegion(reg aws.Region) {
	aws.Region[reg.Name] = reg
}


// Store puts an `ExportResult` struct to an S3 bucket within the specified directory
func (x *S3) Store(result *ExportResult, directory string) *Error {

	if result.Error != nil {
		return result.Error
	}

	file, err := os.Open(result.Path)
	if err != nil {
		return makeErr(err, "")
	}
	defer file.Close()

	if x.UseAwsCli {
		cmd := exec.Command("aws", "s3", "cp", result.Path, "s3://"+x.Bucket+"/")

		var envs []string
		envs = append(envs, "AWS_ACCESS_KEY_ID="+x.AccessKey)
		envs = append(envs, "AWS_SECRET_ACCESS_KEY="+x.ClientSecret)
		envs = append(envs, "AWS_DEFAULT_REGION="+x.Region)

		cmd.Env = envs

		out, err := cmd.CombinedOutput()
		//fmt.Println(cmd)

		if err != nil {
			fmt.Println(string(out))
		}
	} else {
		buffy := bufio.NewReader(file)
		stat, err := file.Stat()
		if err != nil {
			return makeErr(err, "")
		}

		size := stat.Size()

		auth := aws.Auth{
			AccessKey: x.AccessKey,
			SecretKey: x.ClientSecret,
		}

		s := s3.New(auth, aws.Regions[x.Region])
		bucket := s.Bucket(x.Bucket)

		err = bucket.PutReader(directory+result.Filename(), buffy, size, result.MIME, s3.BucketOwnerFull)
	}

	return makeErr(err, "")
}
