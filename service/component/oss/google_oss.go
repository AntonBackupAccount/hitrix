package oss

import (
	"encoding/base64"
	"errors"
	"github.com/coretrix/hitrix/pkg/entity"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/latolukasz/beeorm"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/coretrix/hitrix/pkg/helper"
	"github.com/coretrix/hitrix/service/component/clock"
	"github.com/coretrix/hitrix/service/component/config"
)

type GoogleOSS struct {
	client       *storage.Client
	clockService clock.IClock
	ctx          context.Context
	buckets      map[Bucket]*BucketConfig
	accessID     string
	privateKey   []byte
}

func NewGoogleOSS(configService config.IConfig, clockService clock.IClock, publicNamespaces, privateNamespaces []Namespace) (IProvider, error) {
	ctx := context.Background()

	if !helper.ExistsInDir(".oss.json", configService.GetFolderPath()) {
		return nil, errors.New(configService.GetFolderPath() + "/.oss.json does not exists")
	}

	credentialsFile := configService.GetFolderPath() + "/.oss.json"

	client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))

	if err != nil {
		return nil, err
	}

	jwtCredentialsJSONString, err := os.ReadFile(credentialsFile)

	if err != nil {
		return nil, err
	}

	jwtConfig, _ := google.JWTConfigFromJSON(jwtCredentialsJSONString)

	return &GoogleOSS{
		client:       client,
		clockService: clockService,
		ctx:          ctx,
		buckets:      loadBucketsConfig(configService, publicNamespaces, privateNamespaces),
		accessID:     jwtConfig.Email,
		privateKey:   jwtConfig.PrivateKey,
	}, nil
}

func (ossStorage *GoogleOSS) GetBucketConfig(bucket Bucket) *BucketConfig {
	return getBucketConfig(ossStorage.buckets[bucket])
}

func (ossStorage *GoogleOSS) GetClient() interface{} {
	return ossStorage.client
}

func (ossStorage *GoogleOSS) GetObjectURL(bucket Bucket, object *entity.FileObject) (string, error) {
	cdnURL, err := ossStorage.GetObjectCDNURL(bucket, object)

	if err != nil {
		return "", err
	}

	if cdnURL != "" {
		return cdnURL, nil
	}

	return ossStorage.GetObjectOSSURL(bucket, object)
}

func (ossStorage *GoogleOSS) GetObjectOSSURL(bucket Bucket, object *entity.FileObject) (string, error) {
	ossBucketObjectAttributes, err := ossStorage.client.Bucket(ossStorage.buckets[bucket].Name).Object(object.StorageKey).Attrs(ossStorage.ctx)

	if err != nil {
		return "", err
	}

	return ossBucketObjectAttributes.MediaLink, nil
}

func (ossStorage *GoogleOSS) GetObjectCDNURL(bucket Bucket, object *entity.FileObject) (string, error) {
	return getObjectCDNURL(ossStorage.buckets[bucket], object.StorageKey), nil
}

func (ossStorage *GoogleOSS) GetObjectSignedURL(bucket Bucket, object *entity.FileObject, expires time.Time) (string, error) {
	return storage.SignedURL(ossStorage.buckets[bucket].Name,
		object.StorageKey,
		&storage.SignedURLOptions{
			GoogleAccessID: ossStorage.accessID,
			PrivateKey:     ossStorage.privateKey,
			Method:         http.MethodGet,
			Expires:        expires,
		})
}

func (ossStorage *GoogleOSS) GetObjectBase64Content(bucket Bucket, object *entity.FileObject) (string, error) {
	reader, err := ossStorage.client.Bucket(ossStorage.buckets[bucket].Name).Object(object.StorageKey).NewReader(context.Background())

	if err != nil {
		return "", err
	}

	content, err := io.ReadAll(reader)

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

func (ossStorage *GoogleOSS) UploadObjectFromFile(ormService *beeorm.Engine, bucket Bucket, namespace Namespace, localFile string) (entity.FileObject, error) {
	fileContent, ext, err := readContentFile(localFile)

	if err != nil {
		return entity.FileObject{}, err
	}

	return ossStorage.UploadObjectFromByte(ormService, bucket, namespace, fileContent, ext)
}

func (ossStorage *GoogleOSS) UploadObjectFromBase64(
	ormService *beeorm.Engine,
	bucket Bucket,
	namespace Namespace,
	base64content,
	extension string,
) (entity.FileObject, error) {
	byteData, err := base64.StdEncoding.DecodeString(base64content)

	if err != nil {
		return entity.FileObject{}, err
	}

	return ossStorage.UploadObjectFromByte(ormService, bucket, namespace, byteData, extension)
}

func (ossStorage *GoogleOSS) UploadImageFromBase64(
	ormService *beeorm.Engine,
	bucket Bucket,
	namespace Namespace,
	base64image string,
	extension string,
) (entity.FileObject, error) {
	byteData, err := base64.StdEncoding.DecodeString(base64image)

	if err != nil {
		return entity.FileObject{}, err
	}

	return ossStorage.UploadObjectFromByte(ormService, bucket, namespace, byteData, extension)
}

func (ossStorage *GoogleOSS) UploadImageFromFile(ormService *beeorm.Engine, bucket Bucket, namespace Namespace, localFile string) (entity.FileObject, error) {
	return ossStorage.UploadObjectFromFile(ormService, bucket, namespace, localFile)
}

func (ossStorage *GoogleOSS) UploadObjectFromByte(
	ormService *beeorm.Engine,
	bucket Bucket,
	namespace Namespace,
	objectContent []byte,
	extension string,
) (entity.FileObject, error) {
	bucketConfig := ossStorage.buckets[bucket]

	bucketConfig.validateNamespace(namespace)

	storageCounter := getStorageCounter(ormService, ossStorage.buckets[bucket])

	objectKey := ossStorage.getObjectKey(namespace, storageCounter, extension)

	ossBucketObject := ossStorage.client.Bucket(bucketConfig.Name).Object(objectKey).NewWriter(ossStorage.ctx)

	//TODO Remove
	ossStorage.setObjectContentType(ossBucketObject, extension)

	_, err := ossBucketObject.Write(objectContent)

	if err != nil {
		return entity.FileObject{}, err
	}

	err = ossBucketObject.Close()

	if err != nil {
		return entity.FileObject{}, err
	}

	return entity.FileObject{
		ID:         storageCounter,
		StorageKey: objectKey,
	}, nil
}

func (ossStorage *GoogleOSS) DeleteObject(_ Bucket, _ *entity.FileObject) error {
	panic("not implemented")
}

func (ossStorage *GoogleOSS) CreateObjectFromKey(ormService *beeorm.Engine, bucket Bucket, key string) entity.FileObject {
	return entity.FileObject{
		ID:         getStorageCounter(ormService, ossStorage.buckets[bucket]),
		StorageKey: key,
	}
}

func (ossStorage *GoogleOSS) getObjectKey(namespace Namespace, storageCounter uint64, fileExtension string) string {
	if namespace != "" {
		return namespace.String() + "/" + strconv.FormatUint(storageCounter, 10) + fileExtension
	}

	return strconv.FormatUint(storageCounter, 10) + fileExtension
}

// TODO Remove
func (ossStorage *GoogleOSS) setObjectContentType(writer *storage.Writer, extension string) {
	if writer == nil {
		return
	}

	if extension == ".svg" && writer.ObjectAttrs.ContentType == "" {
		writer.ObjectAttrs.ContentType = "image/svg+xml"
	}
}
