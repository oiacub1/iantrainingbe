package dynamodb

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	authDomain "iantraining/internal/domain/auth"
)

type authRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewAuthRepository crea un nuevo repositorio de autenticación
func NewAuthRepository(client *dynamodb.Client, tableName string) authDomain.Repository {
	return &authRepository{
		client:    client,
		tableName: tableName,
	}
}

// Credenciales item structure para DynamoDB
type credentialsItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	GSI2PK    string `dynamodbav:"GSI2PK"`
	UserID    string `dynamodbav:"userId"`
	Email     string `dynamodbav:"email"`
	Password  string `dynamodbav:"password"`
	CreatedAt int64  `dynamodbav:"createdAt"`
	UpdatedAt int64  `dynamodbav:"updatedAt"`
}

// Refresh token item structure para DynamoDB
type refreshTokenItem struct {
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	UserID    string `dynamodbav:"userId"`
	TokenID   string `dynamodbav:"tokenId"`
	ExpiresAt int64  `dynamodbav:"expiresAt"`
	CreatedAt int64  `dynamodbav:"createdAt"`
}

// Helper functions para construir PK/SK
func (r *authRepository) credentialsPK(userID string) string {
	return fmt.Sprintf("CREDENTIALS#%s", userID)
}

func (r *authRepository) credentialsSK() string {
	return "METADATA"
}

func (r *authRepository) refreshTokenPK(userID string) string {
	return fmt.Sprintf("REFRESH_TOKEN#%s", userID)
}

func (r *authRepository) refreshTokenSK(tokenID string) string {
	return fmt.Sprintf("TOKEN#%s", tokenID)
}

func (r *authRepository) generateTokenID(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])[:16] // Primeros 16 caracteres del hash
}

// CreateCredentials guarda las credenciales de un usuario
func (r *authRepository) CreateCredentials(ctx context.Context, credentials *authDomain.UserCredentials) error {
	now := time.Now().Unix()

	item := credentialsItem{
		PK:        r.credentialsPK(credentials.UserID),
		SK:        r.credentialsSK(),
		GSI2PK:    credentials.Email, // Para buscar por email
		UserID:    credentials.UserID,
		Email:     credentials.Email,
		Password:  credentials.Password,
		CreatedAt: now,
		UpdatedAt: now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials item: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to create credentials: %w", err)
	}

	return nil
}

// GetCredentialsByEmail obtiene las credenciales por email usando GSI2
func (r *authRepository) GetCredentialsByEmail(ctx context.Context, email string) (*authDomain.UserCredentials, error) {
	// Intentar buscar por email usando GSI2 (email como clave de partición)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI2"),
		KeyConditionExpression: aws.String("GSI2PK = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
		Limit: aws.Int32(1),
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query credentials by email: %w", err)
	}

	// Si GSI2 no devuelve resultados (bug de DynamoDB Local), usar Scan como fallback
	if len(result.Items) == 0 {
		// Scan all items with SK = METADATA (credentials only)
		scanInput := &dynamodb.ScanInput{
			TableName:        aws.String(r.tableName),
			FilterExpression: aws.String("SK = :sk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":sk": &types.AttributeValueMemberS{Value: "METADATA"},
			},
		}

		scanResult, scanErr := r.client.Scan(ctx, scanInput)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan credentials: %w", scanErr)
		}

		// Manually filter by email in code
		var foundItem map[string]types.AttributeValue
		for _, item := range scanResult.Items {
			if emailAttr, ok := item["email"]; ok {
				if emailVal, ok := emailAttr.(*types.AttributeValueMemberS); ok {
					if emailVal.Value == email {
						foundItem = item
						break
					}
				}
			}
		}

		if foundItem == nil {
			return nil, authDomain.ErrUserNotFound
		}

		result.Items = []map[string]types.AttributeValue{foundItem}
	}

	var item credentialsItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials item: %w", err)
	}

	// Debug: Log what we got from DynamoDB
	fmt.Printf("DEBUG: Retrieved credentials - Email: %s, Password: %s, UserID: %s\n", item.Email, item.Password, item.UserID)

	return &authDomain.UserCredentials{
		UserID:    item.UserID,
		Email:     item.Email,
		Password:  item.Password,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}, nil
}

// UpdateCredentials actualiza las credenciales de un usuario
func (r *authRepository) UpdateCredentials(ctx context.Context, credentials *authDomain.UserCredentials) error {
	now := time.Now().Unix()

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.credentialsPK(credentials.UserID)},
			"SK": &types.AttributeValueMemberS{Value: r.credentialsSK()},
		},
		UpdateExpression: aws.String("SET #password = :password, #updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#password":  "password",
			"#updatedAt": "updatedAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":password":  &types.AttributeValueMemberS{Value: credentials.Password},
			":updatedAt": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", now)},
		},
	}

	_, err := r.client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update credentials: %w", err)
	}

	return nil
}

// DeleteCredentials elimina las credenciales de un usuario
func (r *authRepository) DeleteCredentials(ctx context.Context, userID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.credentialsPK(userID)},
			"SK": &types.AttributeValueMemberS{Value: r.credentialsSK()},
		},
	}

	_, err := r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}

// StoreRefreshToken guarda un refresh token
func (r *authRepository) StoreRefreshToken(ctx context.Context, userID, tokenID string, expiresAt int64) error {
	now := time.Now().Unix()

	item := refreshTokenItem{
		PK:        r.refreshTokenPK(userID),
		SK:        r.refreshTokenSK(tokenID),
		UserID:    userID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token item: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
	}

	_, err = r.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken obtiene un refresh token por su ID
func (r *authRepository) GetRefreshToken(ctx context.Context, tokenID string) (string, error) {
	// Buscar por tokenID usando GSI1
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :tokenId"),
		FilterExpression:       aws.String("begins_with(PK, :pkPrefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tokenId":  &types.AttributeValueMemberS{Value: tokenID},
			":pkPrefix": &types.AttributeValueMemberS{Value: "REFRESH_TOKEN#"},
		},
		Limit: aws.Int32(1),
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to query refresh token: %w", err)
	}

	if len(result.Items) == 0 {
		return "", authDomain.ErrInvalidRefresh
	}

	var item refreshTokenItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return "", fmt.Errorf("failed to unmarshal refresh token item: %w", err)
	}

	// Verificar que no haya expirado
	if time.Now().Unix() > item.ExpiresAt {
		// Auto-limpieza del token expirado
		r.RevokeRefreshToken(ctx, tokenID)
		return "", authDomain.ErrTokenExpired
	}

	return item.UserID, nil
}

// RevokeRefreshToken revoca un refresh token específico
func (r *authRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	// Primero obtener el userID del token
	userID, err := r.GetRefreshToken(ctx, tokenID)
	if err != nil {
		// Si no existe o está expirado, consideramos que ya está revocado
		return nil
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: r.refreshTokenPK(userID)},
			"SK": &types.AttributeValueMemberS{Value: r.refreshTokenSK(tokenID)},
		},
	}

	_, err = r.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens revoca todos los refresh tokens de un usuario
func (r *authRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: r.refreshTokenPK(userID)},
		},
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to query user refresh tokens: %w", err)
	}

	// Eliminar todos los tokens encontrados
	for _, item := range result.Items {
		var tokenItem refreshTokenItem
		if err := attributevalue.UnmarshalMap(item, &tokenItem); err != nil {
			continue // Ignorar errores individuales
		}

		deleteInput := &dynamodb.DeleteItemInput{
			TableName: aws.String(r.tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: r.refreshTokenPK(userID)},
				"SK": &types.AttributeValueMemberS{Value: r.refreshTokenSK(tokenItem.TokenID)},
			},
		}

		r.client.DeleteItem(ctx, deleteInput)
	}

	return nil
}

// Helper function para crear expressions
func expression(expr string) string {
	return expr
}
