package region

import "sync"

var (
	s3Region string = "us-east-1"
	once     sync.Once
)

func SetS3Region(region string) {
	once.Do(func() {
		s3Region = region
	})
}

func GetS3Region() string {
	return s3Region
}
