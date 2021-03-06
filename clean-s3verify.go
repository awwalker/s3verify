/*
 * Minio S3Verify Library for Amazon S3 Compatible Cloud Storage (C) 2016 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"net/url"

	"github.com/minio/minio-go"
)

// cleanObjects - use minio-go to remove any s3verify created objects.
func cleanObjects(config ServerConfig, bucketName string) error {
	message := "CleanUp (Removing Objects):"
	// Spin scanBar
	scanBar(message)
	// Create new minio-go host from config.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	secure := false
	if hostURL.Scheme == "https" {
		secure = true
	}
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	doneCh := make(chan struct{})
	defer close(doneCh)

	// Only remove s3verify created objects.
	objectCh := client.ListObjects(bucketName, "s3verify/", true, doneCh)
	for object := range objectCh {
		// Spin scanBar
		scanBar(message)
		err := client.RemoveObject(bucketName, object.Key)
		if err != nil {
			// Do not stop on errors.
			continue
		}
	}
	printMessage(message, nil)
	return nil
}

// cleanBucket - use minio-go to cleanup any s3verify created buckets.
func cleanBucket(config ServerConfig, bucketName string) error {
	message := "CleanUp (Removing Buckets):"
	// Spin scanBar
	scanBar(message)
	// Create new minio-go host from config.
	hostURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return err
	}
	secure := false
	if hostURL.Scheme == "https" {
		secure = true
	}
	client, err := minio.New(hostURL.Host, config.Access, config.Secret, secure)
	if err != nil {
		return err
	}
	if err := client.RemoveBucket(bucketName); err != nil {
		return err
	}
	printMessage(message, nil)
	return nil
}

// cleanS3verify - purges the given bucketName of objects then removes the bucket.
func cleanS3verify(config ServerConfig, bucketName string) error {
	if err := cleanObjects(config, bucketName); err != nil {
		return err
	}
	if err := cleanBucket(config, bucketName); err != nil {
		return err
	}
	return nil
}
