package pulumiservice

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-azure-native-sdk/containerregistry/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/qbeast-io/provider-acr-cache/apis/containerregistry/v1alpha1"
	"github.com/qbeast-io/provider-acr-cache/internal/controller/config"
	"os"
	"strings"
)

type Service struct {
	credentials *config.AzureCredentials
}

func NewService(creds []byte) (*Service, error) {
	credentials := &config.AzureCredentials{}
	err := json.Unmarshal(creds, credentials)
	if err != nil {
		return nil, err
	}
	return &Service{credentials}, nil
}

func (s *Service) ApplyCredentialSet(ctx context.Context, spec *v1alpha1.CredentialSetSpec) (*v1alpha1.CredentialSetObservation, error) {
	stackName := fmt.Sprintf("credset-%s", spec.ForProvider.Name)

	if spec.ForProvider.ResourceGroupName == "" {
		return nil, errors.New("ResourceGroupName is required")
	}
	if spec.ForProvider.RegistryName == "" {
		return nil, errors.New("RegistryName is required")
	}
	if spec.ForProvider.Name == "" {
		return nil, errors.New("Name is required")
	}
	credentialSet := spec.ForProvider
	stack, err := auto.UpsertStackInlineSource(ctx, stackName, "", func(ctx *pulumi.Context) error {
		credentialArray := make([]containerregistry.AuthCredentialInput, len(credentialSet.AuthCredentials))
		for i, credential := range credentialSet.AuthCredentials {
			credentialArray[i] = containerregistry.AuthCredentialArgs{
				Name:                     pulumi.String(credential.Name),
				PasswordSecretIdentifier: pulumi.String(credential.PasswordSecretIdentifier),
				UsernameSecretIdentifier: pulumi.String(credential.UsernameSecretIdentifier),
			}
		}
		var identityType containerregistry.ResourceIdentityType
		userAssignedIdentities := make(map[string]containerregistry.UserIdentityPropertiesInput)
		switch credentialSet.Identity.Type {
		case "SystemAssigned":
			identityType = containerregistry.ResourceIdentityTypeSystemAssigned
		case "UserAssigned":
			identityType = containerregistry.ResourceIdentityTypeUserAssigned
			setUserIdentitiesFromSpec(credentialSet, &userAssignedIdentities)
		case "SystemAssigned, UserAssigned":
			identityType = containerregistry.ResourceIdentityType_SystemAssigned_UserAssigned
			setUserIdentitiesFromSpec(credentialSet, &userAssignedIdentities)
		case "None":
			identityType = containerregistry.ResourceIdentityTypeNone
		}
		_, err := containerregistry.NewCredentialSet(ctx, "credentialSet", &containerregistry.CredentialSetArgs{
			AuthCredentials:   containerregistry.AuthCredentialArray(credentialArray),
			CredentialSetName: pulumi.String(credentialSet.Name),
			Identity: containerregistry.IdentityPropertiesArgs{
				Type:                   identityType,
				UserAssignedIdentities: containerregistry.UserIdentityPropertiesMap(userAssignedIdentities),
			},
			LoginServer:       pulumi.String(credentialSet.LoginServer),
			RegistryName:      pulumi.String(credentialSet.RegistryName),
			ResourceGroupName: pulumi.String(credentialSet.ResourceGroupName),
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	err = stack.SetAllConfig(ctx, map[string]auto.ConfigValue{
		"azure:clientId":       {Value: s.credentials.ClientId},
		"azure:clientSecret":   {Value: s.credentials.ClientSecret},
		"azure:tenantId":       {Value: s.credentials.TenantId},
		"azure:subscriptionId": {Value: s.credentials.SubscriptionId},
	})
	if err != nil {
		return nil, err
	}

	res, err := stack.Up(ctx, optup.ProgressStreams(os.Stdout))
	if err != nil {
		return nil, err
	}

	systemData := res.Outputs["systemData"].Value.(map[string]interface{})
	identity := res.Outputs["identity"].Value.(map[string]interface{})
	createdByType := systemData["createdByType"].(string)
	createdAt := systemData["createdAt"].(string)
	createdBy := systemData["createdBy"].(string)
	lastModifiedByType := systemData["lastModifiedByType"].(string)
	lastModifiedAt := systemData["lastModifiedAt"].(string)
	lastModifiedBy := systemData["lastModifiedBy"].(string)
	identityPrincipalId := identity["principalId"].(string)
	identityTenantId := identity["tenantId"].(string)
	identityType := identity["type"].(string)
	return &v1alpha1.CredentialSetObservation{
		AzureApiVersion:     res.Outputs["azureApiVersion"].Value.(string),
		CreationDate:        res.Outputs["creationDate"].Value.(string),
		Id:                  res.Outputs["id"].Value.(string),
		Name:                res.Outputs["name"].Value.(string),
		ProvisioningState:   res.Outputs["provisioningState"].Value.(string),
		Type:                res.Outputs["type"].Value.(string),
		IdentityPrincipalId: identityPrincipalId,
		IdentityTenantId:    identityTenantId,
		IdentityType:        identityType,
		CreatedByType:       createdByType,
		CreatedAt:           createdAt,
		CreatedBy:           createdBy,
		LastModifiedByType:  lastModifiedByType,
		LastModifiedAt:      lastModifiedAt,
		LastModifiedBy:      lastModifiedBy,
		Ready:               res.Outputs["provisioningState"].Value.(string) == "Succeeded",
	}, nil
}

func (s *Service) ObserveCredentialSet(ctx context.Context, spec *v1alpha1.CredentialSetSpec) (*v1alpha1.CredentialSetObservation, bool /* exists */, bool /* upToDate */, error) {
	creds := spec.ForProvider
	var obs *v1alpha1.CredentialSetObservation
	var exists bool
	var upToDate bool
	var err error
	pulumi.Run(func(pctx *pulumi.Context) error {
		cred, err := containerregistry.LookupCredentialSet(pctx, &containerregistry.LookupCredentialSetArgs{
			CredentialSetName: creds.Name,
			ResourceGroupName: creds.ResourceGroupName,
			RegistryName:      creds.RegistryName,
		})
		if err != nil {
			if strings.Contains(err.Error(), "was not found") {
				exists = false
				return nil
			}
			err = errors.Wrap(err, "failed to lookup credential set")
			return nil
		}
		obs = &v1alpha1.CredentialSetObservation{
			Id:                  cred.Id,
			Name:                cred.Name,
			ProvisioningState:   cred.ProvisioningState,
			CreationDate:        cred.CreationDate,
			Type:                cred.Type,
			IdentityType:        *cred.Identity.Type,
			IdentityTenantId:    *cred.Identity.TenantId,
			IdentityPrincipalId: *cred.Identity.PrincipalId,
			Ready:               cred.ProvisioningState == "Succeeded",
			CreatedByType:       *cred.SystemData.CreatedByType,
			CreatedAt:           *cred.SystemData.CreatedAt,
			CreatedBy:           *cred.SystemData.CreatedBy,
			LastModifiedByType:  *cred.SystemData.LastModifiedByType,
			LastModifiedAt:      *cred.SystemData.LastModifiedAt,
			LastModifiedBy:      *cred.SystemData.LastModifiedBy,
		}
		cred.ProvisioningState = cred.ProvisioningState
		upToDate = credentialSetIsUpToDate(&spec.ForProvider, cred)
		return nil
	})
	return obs, exists, upToDate, err
}

func (s *Service) DeleteCredentialSet(ctx context.Context, spec *v1alpha1.CredentialSetSpec) error {
	stackName := fmt.Sprintf("credset-%s", spec.ForProvider.Name)

	stack, err := auto.SelectStackInlineSource(ctx, stackName, "", func(ctx *pulumi.Context) error {
		return nil
	})
	if err != nil {
		if auto.IsSelectStack404Error(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to select Pulumi stack: %w", err)
	}

	// Run destroy
	_, err = stack.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("failed to destroy stack: %w", err)
	}

	// Optionally delete the stack from the backend
	err = stack.Workspace().RemoveStack(ctx, stackName)
	if err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}

	return nil
}

func (s *Service) ApplyCacheRule(ctx context.Context, spec *v1alpha1.CacheRuleSpec) (*v1alpha1.CacheRuleObservation, error) {
	stackName := fmt.Sprintf("credset-%s", spec.ForProvider.CacheRuleName)
	if spec.ForProvider.CacheRuleName == "" {
		return nil, errors.New("CacheRuleName is required")
	}
	if spec.ForProvider.ResourceGroupName == "" {
		return nil, errors.New("ResourceGroupName is required")
	}
	if spec.ForProvider.RegistryName == "" {
		return nil, errors.New("RegistryName is required")
	}
	if spec.ForProvider.SourceRepository == "" {
		return nil, errors.New("SourceRepository is required")
	}
	if spec.ForProvider.TargetRepository == "" {
		return nil, errors.New("TargetRepository is required")
	}
	if spec.ForProvider.CredentialSetResourceId == "" && spec.ForProvider.CredentialSetName == "" {
		return nil, errors.New("CredentialSetResourceId or CredentialSetName is required")
	}
	// If CredentialSetResourceId is not provided, look it up using CredentialSetName
	if spec.ForProvider.CredentialSetResourceId == "" {
		obs, exists, _, err := s.ObserveCredentialSet(ctx, &v1alpha1.CredentialSetSpec{
			ForProvider: v1alpha1.CredentialSetParameters{
				Name:              spec.ForProvider.CredentialSetName,
				RegistryName:      spec.ForProvider.RegistryName,
				ResourceGroupName: spec.ForProvider.ResourceGroupName,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to lookup credential set")
		}
		if !exists {
			return nil, errors.New("credential set does not exist")
		}
		spec.ForProvider.CredentialSetResourceId = obs.Id
	}
	stack, err := auto.UpsertStackInlineSource(ctx, stackName, "", func(ctx *pulumi.Context) error {
		_, err := containerregistry.NewCacheRule(ctx, "cacheRule", &containerregistry.CacheRuleArgs{
			CacheRuleName:           pulumi.String(spec.ForProvider.CacheRuleName),
			CredentialSetResourceId: pulumi.String(spec.ForProvider.CredentialSetResourceId),
			RegistryName:            pulumi.String(spec.ForProvider.RegistryName),
			ResourceGroupName:       pulumi.String(spec.ForProvider.ResourceGroupName),
			SourceRepository:        pulumi.String(spec.ForProvider.SourceRepository),
			TargetRepository:        pulumi.String(spec.ForProvider.TargetRepository),
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	err = stack.SetAllConfig(ctx, map[string]auto.ConfigValue{
		"azure:clientId":       {Value: s.credentials.ClientId},
		"azure:clientSecret":   {Value: s.credentials.ClientSecret},
		"azure:tenantId":       {Value: s.credentials.TenantId},
		"azure:subscriptionId": {Value: s.credentials.SubscriptionId},
	})
	if err != nil {
		return nil, err
	}

	res, err := stack.Up(ctx, optup.ProgressStreams(os.Stdout))
	if err != nil {
		return nil, err
	}

	systemData := res.Outputs["systemData"].Value.(map[string]interface{})
	createdByType := systemData["createdByType"].(string)
	createdAt := systemData["createdAt"].(string)
	createdBy := systemData["createdBy"].(string)
	lastModifiedByType := systemData["lastModifiedByType"].(string)
	lastModifiedAt := systemData["lastModifiedAt"].(string)
	lastModifiedBy := systemData["lastModifiedBy"].(string)
	return &v1alpha1.CacheRuleObservation{
		AzureApiVersion:    res.Outputs["azureApiVersion"].Value.(string),
		CreationDate:       res.Outputs["creationDate"].Value.(string),
		Id:                 res.Outputs["id"].Value.(string),
		Name:               res.Outputs["name"].Value.(string),
		ProvisioningState:  res.Outputs["provisioningState"].Value.(string),
		Type:               res.Outputs["type"].Value.(string),
		CreatedByType:      createdByType,
		CreatedAt:          createdAt,
		CreatedBy:          createdBy,
		LastModifiedByType: lastModifiedByType,
		LastModifiedAt:     lastModifiedAt,
		LastModifiedBy:     lastModifiedBy,
		Ready:              res.Outputs["provisioningState"].Value.(string) == "Succeeded",
	}, nil
}

func (s *Service) ObserveCacheRule(ctx context.Context, spec *v1alpha1.CacheRuleSpec) (*v1alpha1.CacheRuleObservation, bool /* exists */, bool /* upToDate */, error) {
	var obs *v1alpha1.CacheRuleObservation
	var exists bool
	var upToDate bool
	var err error
	pulumi.Run(func(pctx *pulumi.Context) error {
		cred, err := containerregistry.LookupCacheRule(pctx, &containerregistry.LookupCacheRuleArgs{
			CacheRuleName:     spec.ForProvider.CacheRuleName,
			RegistryName:      spec.ForProvider.RegistryName,
			ResourceGroupName: spec.ForProvider.ResourceGroupName,
		})
		if err != nil {
			if strings.Contains(err.Error(), "was not found") {
				exists = false
				return nil
			}
			err = errors.Wrap(err, "failed to lookup credential set")
			return nil
		}
		obs = &v1alpha1.CacheRuleObservation{
			CreationDate:       cred.CreationDate,
			Id:                 cred.Id,
			Name:               cred.Name,
			ProvisioningState:  cred.ProvisioningState,
			Type:               cred.Type,
			CreatedByType:      *cred.SystemData.CreatedByType,
			CreatedAt:          *cred.SystemData.CreatedAt,
			CreatedBy:          *cred.SystemData.CreatedBy,
			LastModifiedByType: *cred.SystemData.LastModifiedByType,
			LastModifiedAt:     *cred.SystemData.LastModifiedAt,
			LastModifiedBy:     *cred.SystemData.LastModifiedBy,
			Ready:              cred.ProvisioningState == "Succeeded",
		}
		upToDate = cacheRuleIsUpToDate(&spec.ForProvider, cred)
		return nil
	})
	return obs, exists, upToDate, err
}

func cacheRuleIsUpToDate(c *v1alpha1.CacheRuleParameters, rule *containerregistry.LookupCacheRuleResult) bool {
	if c.CacheRuleName != rule.Name {
		return false
	}
	if c.CredentialSetResourceId != *rule.CredentialSetResourceId {
		return false
	}
	if c.SourceRepository != *rule.SourceRepository {
		return false
	}
	if c.TargetRepository != *rule.TargetRepository {
		return false
	}
	return true
}

func (s *Service) DeleteCacheRule(ctx context.Context, spec *v1alpha1.CacheRuleSpec) error {
	stackName := fmt.Sprintf("credset-%s", spec.ForProvider.CacheRuleName)

	stack, err := auto.SelectStackInlineSource(ctx, stackName, "", func(ctx *pulumi.Context) error {
		return nil
	})
	if err != nil {
		if auto.IsSelectStack404Error(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to select Pulumi stack: %w", err)
	}

	// Run destroy
	_, err = stack.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("failed to destroy stack: %w", err)
	}

	// Optionally delete the stack from the backend
	err = stack.Workspace().RemoveStack(ctx, stackName)
	if err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}

	return nil
}

func credentialSetIsUpToDate(c *v1alpha1.CredentialSetParameters, cred *containerregistry.LookupCredentialSetResult) bool {
	// Compare the spec with the observed state
	if c.Name != cred.Name {
		return false
	}
	if c.LoginServer != *cred.LoginServer {
		return false
	}
	if c.Identity.Type != *cred.Identity.Type {
		return false
	}
	if c.Identity.PrincipalId != *cred.Identity.PrincipalId {
		return false
	}
	if c.Identity.TenantId != *cred.Identity.TenantId {
		return false
	}
	for k, v := range c.Identity.UserAssignedIdentities {
		if *cred.Identity.UserAssignedIdentities[k].ClientId != v.ClientId || *cred.Identity.UserAssignedIdentities[k].PrincipalId != v.PrincipalId {
			return false
		}
	}
	for _, credential := range c.AuthCredentials {
		for _, cred := range cred.AuthCredentials {
			if credential.Name == *cred.Name {
				if credential.UsernameSecretIdentifier != *cred.UsernameSecretIdentifier || credential.PasswordSecretIdentifier != *cred.PasswordSecretIdentifier {
					return false
				}
			}
		}
	}
	return true
}

func setUserIdentitiesFromSpec(spec v1alpha1.CredentialSetParameters, identities *map[string]containerregistry.UserIdentityPropertiesInput) {
	for k, v := range spec.Identity.UserAssignedIdentities {
		(*identities)[k] = containerregistry.UserIdentityPropertiesArgs{
			ClientId:    pulumi.String(v.ClientId),
			PrincipalId: pulumi.String(v.PrincipalId),
		}
	}
}
