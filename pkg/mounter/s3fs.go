package mounter

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/yandex-cloud/k8s-csi-s3/pkg/s3"
)

// Implements Mounter
type s3fsMounter struct {
	meta         *s3.FSMeta
	url          string
	region       string
	ak           string
	sk           string
	sessionToken string
}

const (
	s3fsCmd = "s3fs"
)

func newS3fsMounter(meta *s3.FSMeta, cfg *s3.Config) (Mounter, error) {
	return &s3fsMounter{
		meta: meta,
		// url:           cfg.Endpoint,
		// region:        cfg.Region,
		url:          "https://uat-s3.siliconflow.cn",
		region:       "oss-cn-shanghai",
		ak:           cfg.AccessKeyID,
		sk:           cfg.SecretAccessKey,
		sessionToken: cfg.SessionToken,
	}, nil
}

func (s3fs *s3fsMounter) Mount(target, volumeID string) error {
	glog.Infof("--> Mounting S3FS volume: %s at target: %s", volumeID, target)

	// cancel pwFile method, use env vars
	// if err := writes3fsPass(s3fs.pwFileContent); err != nil {
	// 	return err
	// }

	envVars := []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3fs.ak),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3fs.sk),
		// AWS_SESSION_TOKEN
	}
	if s3fs.sessionToken != "" {
		envVars = append(envVars, fmt.Sprintf("AWS_SESSION_TOKEN=%s", s3fs.sessionToken))
	}

	glog.Infof("--> Mounting S3FS args, bucket: %s prefix: %s", s3fs.meta.BucketName, s3fs.meta.Prefix)

	args := []string{
		fmt.Sprintf("%s:/%s", s3fs.meta.BucketName, s3fs.meta.Prefix),
		target,
		"-o", fmt.Sprintf("url=%s", s3fs.url),
		"-o", fmt.Sprintf("endpoint=%s", s3fs.region),
		"-o", "allow_other",
		"-o", "mp_umask=000",
		// debug options
		"-o", "use_path_request_style",
		"-o", "dbglevel=info",
		"-o", "curldbg",
		"-o", "no_check_certificate",
		"-d",
	}

	// args = append(args, s3fs.meta.MountOptions...)
	return fuseMount(target, s3fsCmd, args, envVars)
}

func writes3fsPass(pwFileContent string) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3fs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
