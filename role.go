package main

import (
	"context"
	"fmt"

	"github.com/yaegashi/customazed/utils/azutil"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-09-01-preview/authorization"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const (
	RoleNameStorageBlobDataContributor = "ba92f5b4-2d11-453d-a403-e96b0029c9fe"
	RoleNameStorageBlobDataOwner       = "b7e6dc6d-f1e8-4753-8033-0f276bb0955b"
	RoleNameStorageBlobDataReader      = "2a2b9908-6ea1-4ae2-8e65-a410df84e7d1"
	RoleNameStorageBlobDataDelegator   = "db58b8e5-c6ad-4a2a-8342-4190687cbf4a"
	RoleNameImageCreator               = "d3d5cf35-0954-4711-b01a-faa4800979d5"
)

func (app *App) RoleID(name string) *string {
	s := fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", app.Config.SubscriptionID, name)
	return &s
}

func (app *App) RoleSetup(ctx context.Context) error {
	container, err := app.StorageContainer(ctx)
	if err != nil {
		return err
	}

	identity, err := app.Identity(ctx)
	if err != nil {
		return err
	}

	machine, err := app.Machine(ctx)
	if err != nil {
		return err
	}

	image, err := app.Image(ctx)
	if err != nil {
		return err
	}

	gallery, err := app.Gallery(ctx)
	if err != nil {
		return err
	}

	authorizer, err := app.ARMAuthorizer()
	if err != nil {
		return err
	}

	roleAssignmentsClient := authorization.NewRoleAssignmentsClient(app.Config.SubscriptionID)
	roleAssignmentsClient.Authorizer = authorizer

	if container != nil {
		storageToken, err := app.StorageToken()
		if err != nil {
			return err
		}
		parser, claims := &jwt.Parser{}, jwt.MapClaims{}
		_, _, err = parser.ParseUnverified(storageToken.OAuthToken(), claims)
		if err != nil {
			return err
		}
		if oid, ok := claims["oid"].(string); ok {
			app.Logf("Role: assign role to user for blob container")
			roleAssignmentParams := authorization.RoleAssignmentCreateParameters{
				RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
					RoleDefinitionID: app.RoleID(RoleNameStorageBlobDataOwner),
					PrincipalID:      &oid,
				},
			}
			_, err = roleAssignmentsClient.Create(ctx, *container.ID, uuid.New().String(), roleAssignmentParams)
			if err != nil {
				if aErr := azutil.Error(err); aErr == nil || aErr.ServiceError.Code != "RoleAssignmentExists" {
					return err
				}
			}
		}
		if identity != nil {
			app.Logf("Role: assign role to identity for blob container")
			roleAssignmentParams := authorization.RoleAssignmentCreateParameters{
				RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
					RoleDefinitionID: app.RoleID(RoleNameStorageBlobDataReader),
					PrincipalID:      to.StringPtr(identity.PrincipalID.String()),
				},
			}
			_, err = roleAssignmentsClient.Create(ctx, *container.ID, uuid.New().String(), roleAssignmentParams)
			if err != nil {
				if aErr := azutil.Error(err); aErr == nil || aErr.ServiceError.Code != "RoleAssignmentExists" {
					return err
				}
			}
		}
		if machine != nil {
			app.Logf("Role: assign role to machine for blob container")
			roleAssignmentParams := authorization.RoleAssignmentCreateParameters{
				RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
					RoleDefinitionID: app.RoleID(RoleNameStorageBlobDataReader),
					PrincipalID:      machine.Identity.PrincipalID,
				},
			}
			_, err = roleAssignmentsClient.Create(ctx, *container.ID, uuid.New().String(), roleAssignmentParams)
			if err != nil {
				if aErr := azutil.Error(err); aErr == nil || aErr.ServiceError.Code != "RoleAssignmentExists" {
					return err
				}
			}
		}
	}

	if identity != nil {
		app.Logf("Role: creating custom role for identity")
		defintionsClient := authorization.NewRoleDefinitionsClient(app.Config.SubscriptionID)
		defintionsClient.Authorizer = authorizer
		subscriptionScope := fmt.Sprintf("/subscriptions/%s", app.Config.SubscriptionID)
		definition := authorization.RoleDefinition{
			RoleDefinitionProperties: &authorization.RoleDefinitionProperties{
				RoleName:         to.StringPtr("Azure Image Builder Service Image Creation Role"),
				Description:      to.StringPtr("Azure Image Builder Service access to image resources"),
				AssignableScopes: &[]string{subscriptionScope},
				Permissions: &[]authorization.Permission{
					{
						Actions: &[]string{
							"Microsoft.Compute/galleries/read",
							"Microsoft.Compute/galleries/images/read",
							"Microsoft.Compute/galleries/images/versions/read",
							"Microsoft.Compute/galleries/images/versions/write",
							"Microsoft.Compute/images/write",
							"Microsoft.Compute/images/read",
							"Microsoft.Compute/images/delete",
						},
					},
				},
			},
		}
		definition, err = defintionsClient.CreateOrUpdate(ctx, subscriptionScope, RoleNameImageCreator, definition)
		if err != nil {
			return err
		}
		roleAssignmentParams := authorization.RoleAssignmentCreateParameters{
			RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: definition.ID,
				PrincipalID:      to.StringPtr(identity.PrincipalID.String()),
			},
		}
		if image != nil {
			app.Logf("Role: assign role to identity for image")
			groupScope := fmt.Sprintf("%s/resourceGroups/%s", subscriptionScope, app.Config.Image.ResourceGroup)
			_, err = roleAssignmentsClient.Create(ctx, groupScope, uuid.New().String(), roleAssignmentParams)
			if err != nil {
				if aErr := azutil.Error(err); aErr == nil || aErr.ServiceError.Code != "RoleAssignmentExists" {
					return err
				}
			}
		}
		if gallery != nil {
			app.Logf("Role: assign role to identity for gallery")
			_, err = roleAssignmentsClient.Create(ctx, *gallery.ID, uuid.New().String(), roleAssignmentParams)
			if err != nil {
				if aErr := azutil.Error(err); aErr == nil || aErr.ServiceError.Code != "RoleAssignmentExists" {
					return err
				}
			}
		}
	}

	return nil
}
