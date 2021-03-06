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
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Store all objects that are uploaded by s3verify tests.
var s3verifyObjects = []*ObjectInfo{}

// Store all objects that are uploaded through the preparing operation.
var preparedObjects = []*ObjectInfo{}

// Store all objects that were copied.
var copyObjects = []*ObjectInfo{}

// newPutObjectReq - Create a new HTTP request for PUT object.
func newPutObjectReq(bucketName, objectName string, objectData []byte) (Request, error) {
	// An HTTP request for a PUT object.
	var putObjectReq = Request{
		customHeader: http.Header{},
	}

	// Set the bucketName and objectName.
	putObjectReq.bucketName = bucketName
	putObjectReq.objectName = objectName

	// Compute md5Sum and sha256Sum from the input data.
	reader := bytes.NewReader(objectData)
	md5Sum, sha256Sum, contentLength, err := computeHash(reader)
	if err != nil {
		return Request{}, err
	}

	putObjectReq.customHeader.Set("Content-MD5", base64.StdEncoding.EncodeToString(md5Sum))
	putObjectReq.customHeader.Set("X-Amz-Content-Sha256", hex.EncodeToString(sha256Sum))
	putObjectReq.customHeader.Set("User-Agent", appUserAgent)

	putObjectReq.contentLength = contentLength
	// Set the body to the data held in objectData.
	putObjectReq.contentBody = reader

	return putObjectReq, nil
}

// putObjectVerify - Verify the response matches what is expected.
func putObjectVerify(res *http.Response, expectedStatusCode int) error {
	if err := verifyHeaderPutObject(res.Header); err != nil {
		return err
	}
	if err := verifyStatusPutObject(res.StatusCode, expectedStatusCode); err != nil {
		return err
	}
	if err := verifyBodyPutObject(res.Body); err != nil {
		return err
	}
	return nil
}

// verifyStatusPutObject - Verify that the res.StatusCode code matches what is expected.
func verifyStatusPutObject(respStatusCode, expectedStatusCode int) error {
	if respStatusCode != expectedStatusCode {
		err := fmt.Errorf("Unexpected Response Status Code: wanted %v, got %v", expectedStatusCode, respStatusCode)
		return err
	}
	return nil
}

// verifyBodyPutObject - Verify that the body returned matches what is uploaded.
func verifyBodyPutObject(resBody io.Reader) error {
	body, err := ioutil.ReadAll(resBody)
	if err != nil {
		return err
	}
	// A PUT request should give back an empty body.
	if !bytes.Equal(body, []byte{}) {
		err := fmt.Errorf("Unexpected Body Recieved: expected empty body but recieved: %v", string(body))
		return err
	}
	return nil
}

// verifyHeaderPutObject - Verify that the header returned matches what is expected.
func verifyHeaderPutObject(header http.Header) error {
	if err := verifyStandardHeaders(header); err != nil {
		return err
	}
	return nil
}

// TODO: need mainPutObjectPrepared and mainPutObjectUnPrepared.
func mainPutObjectPrepared(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutObject:", curTest, globalTotalNumTest)
	// Use the last bucket created by s3verify itself.
	bucket := s3verifyBuckets[0]
	// Spin scanBar
	scanBar(message)
	// Since the use of --prepare will have set up enough objects for future tests
	// only add one more additional object.
	object := &ObjectInfo{
		Key:  "s3verify/made/put/object",
		Body: []byte(randString(60, rand.NewSource(time.Now().UnixNano()), "")),
	}
	// Spin scanBar
	scanBar(message)
	// Create a new request.
	req, err := newPutObjectReq(bucket.Name, object.Key, object.Body)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	// Execute the request.
	res, err := config.execRequest("PUT", req)
	if err != nil {
		printMessage(message, err)
		return false
	}
	// Spin scanBar
	scanBar(message)
	defer closeResponse(res)
	// Verify the response.
	if err := putObjectVerify(res, http.StatusOK); err != nil {
		printMessage(message, err)
		return false
	}
	// Store this object in the global objects list.
	s3verifyObjects = append(s3verifyObjects, object)
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}

// Test a PUT object request with no special headers set. This adds one object to each of the test buckets.
func mainPutObjectUnPrepared(config ServerConfig, curTest int) bool {
	message := fmt.Sprintf("[%02d/%d] PutObject:", curTest, globalTotalNumTest)
	// TODO: create tests designed to fail.
	bucket := s3verifyBuckets[0]
	// Spin scanBar
	scanBar(message)
	// TODO: need to update to 1001 once this is production ready.
	// Upload 1001 objects with 1 byte each to check the ListObjects API with.
	for i := 0; i < 101; i++ {
		// Spin scanBar
		scanBar(message)
		object := &ObjectInfo{}
		object.Key = "s3verify/put/object/" + strconv.Itoa(i)
		// Create 60 bytes worth of random data for each object.
		body := randString(60, rand.NewSource(time.Now().UnixNano()), "")
		object.Body = []byte(body)
		// Create a new request.
		req, err := newPutObjectReq(bucket.Name, object.Key, object.Body)
		if err != nil {
			printMessage(message, err)
			return false
		}
		// Execute the request.
		res, err := config.execRequest("PUT", req)
		if err != nil {
			printMessage(message, err)
			return false
		}
		defer closeResponse(res)
		// Verify the response.
		if err := putObjectVerify(res, http.StatusOK); err != nil {
			printMessage(message, err)
			return false
		}
		// Add the new object to the list of objects.
		s3verifyObjects = append(s3verifyObjects, object)
		// Spin scanBar
		scanBar(message)
	}
	// Spin scanBar
	scanBar(message)
	// Test passed.
	printMessage(message, nil)
	return true
}
