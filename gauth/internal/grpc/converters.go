package grpc

import (
	pb "github.com/dojima-foundation/tee-auth/gauth/api/proto"
	"github.com/dojima-foundation/tee-auth/gauth/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// convertOrganizationToProto converts a models.Organization to pb.Organization
func convertOrganizationToProto(org *models.Organization) *pb.Organization {
	if org == nil {
		return nil
	}

	// Convert users
	users := make([]*pb.User, len(org.Users))
	for i, user := range org.Users {
		users[i] = convertUserToProto(&user)
	}

	// Convert invitations
	invitations := make([]*pb.Invitation, len(org.Invitations))
	for i, inv := range org.Invitations {
		invitations[i] = convertInvitationToProto(&inv)
	}

	// Convert policies
	policies := make([]*pb.Policy, len(org.Policies))
	for i, policy := range org.Policies {
		policies[i] = convertPolicyToProto(&policy)
	}

	// Convert tags
	tags := make([]*pb.Tag, len(org.Tags))
	for i, tag := range org.Tags {
		tags[i] = convertTagToProto(&tag)
	}

	// Convert private keys
	privateKeys := make([]*pb.PrivateKey, len(org.PrivateKeys))
	for i, pk := range org.PrivateKeys {
		privateKeys[i] = convertPrivateKeyToProto(&pk)
	}

	// Convert wallets
	wallets := make([]*pb.Wallet, len(org.Wallets))
	for i, wallet := range org.Wallets {
		wallets[i] = convertWalletToProto(&wallet)
	}

	return &pb.Organization{
		Id:          org.ID.String(),
		Version:     org.Version,
		Name:        org.Name,
		Users:       users,
		RootQuorum:  convertQuorumToProto(&org.RootQuorum),
		Invitations: invitations,
		Policies:    policies,
		Tags:        tags,
		PrivateKeys: privateKeys,
		Wallets:     wallets,
		CreatedAt:   timestamppb.New(org.CreatedAt),
		UpdatedAt:   timestamppb.New(org.UpdatedAt),
	}
}

// convertUserToProto converts a models.User to pb.User
func convertUserToProto(user *models.User) *pb.User {
	if user == nil {
		return nil
	}

	// Convert auth methods
	authMethods := make([]*pb.AuthMethod, len(user.AuthMethods))
	for i, am := range user.AuthMethods {
		authMethods[i] = convertAuthMethodToProto(&am)
	}

	return &pb.User{
		Id:             user.ID.String(),
		OrganizationId: user.OrganizationID.String(),
		Username:       user.Username,
		Email:          user.Email,
		PublicKey:      user.PublicKey,
		AuthMethods:    authMethods,
		Tags:           user.Tags,
		IsActive:       user.IsActive,
		CreatedAt:      timestamppb.New(user.CreatedAt),
		UpdatedAt:      timestamppb.New(user.UpdatedAt),
	}
}

// convertAuthMethodToProto converts a models.AuthMethod to pb.AuthMethod
func convertAuthMethodToProto(am *models.AuthMethod) *pb.AuthMethod {
	if am == nil {
		return nil
	}

	return &pb.AuthMethod{
		Id:        am.ID.String(),
		UserId:    am.UserID.String(),
		Type:      am.Type,
		Name:      am.Name,
		Data:      am.Data,
		IsActive:  am.IsActive,
		CreatedAt: timestamppb.New(am.CreatedAt),
		UpdatedAt: timestamppb.New(am.UpdatedAt),
	}
}

// convertQuorumToProto converts a models.Quorum to pb.Quorum
func convertQuorumToProto(quorum *models.Quorum) *pb.Quorum {
	if quorum == nil {
		return nil
	}

	userIds := make([]string, len(quorum.UserIDs))
	for i, id := range quorum.UserIDs {
		userIds[i] = id.String()
	}

	return &pb.Quorum{
		UserIds:   userIds,
		Threshold: int32(quorum.Threshold),
	}
}

// convertInvitationToProto converts a models.Invitation to pb.Invitation
func convertInvitationToProto(inv *models.Invitation) *pb.Invitation {
	if inv == nil {
		return nil
	}

	invitation := &pb.Invitation{
		Id:             inv.ID.String(),
		OrganizationId: inv.OrganizationID.String(),
		Email:          inv.Email,
		Role:           inv.Role,
		Token:          inv.Token,
		ExpiresAt:      timestamppb.New(inv.ExpiresAt),
		CreatedAt:      timestamppb.New(inv.CreatedAt),
	}

	if inv.AcceptedAt != nil {
		invitation.AcceptedAt = timestamppb.New(*inv.AcceptedAt)
	}

	return invitation
}

// convertPolicyToProto converts a models.Policy to pb.Policy
func convertPolicyToProto(policy *models.Policy) *pb.Policy {
	if policy == nil {
		return nil
	}

	return &pb.Policy{
		Id:             policy.ID.String(),
		OrganizationId: policy.OrganizationID.String(),
		Name:           policy.Name,
		Description:    policy.Description,
		Rules:          policy.Rules,
		IsActive:       policy.IsActive,
		CreatedAt:      timestamppb.New(policy.CreatedAt),
		UpdatedAt:      timestamppb.New(policy.UpdatedAt),
	}
}

// convertTagToProto converts a models.Tag to pb.Tag
func convertTagToProto(tag *models.Tag) *pb.Tag {
	if tag == nil {
		return nil
	}

	return &pb.Tag{
		Id:             tag.ID.String(),
		OrganizationId: tag.OrganizationID.String(),
		Name:           tag.Name,
		Description:    tag.Description,
		Color:          tag.Color,
		CreatedAt:      timestamppb.New(tag.CreatedAt),
	}
}

// convertPrivateKeyToProto converts a models.PrivateKey to pb.PrivateKey
func convertPrivateKeyToProto(pk *models.PrivateKey) *pb.PrivateKey {
	if pk == nil {
		return nil
	}

	return &pb.PrivateKey{
		Id:             pk.ID.String(),
		OrganizationId: pk.OrganizationID.String(),
		WalletId:       pk.WalletID.String(),
		Name:           pk.Name,
		PublicKey:      pk.PublicKey,
		Curve:          pk.Curve,
		Path:           pk.Path,
		Tags:           pk.Tags,
		IsActive:       pk.IsActive,
		CreatedAt:      timestamppb.New(pk.CreatedAt),
		UpdatedAt:      timestamppb.New(pk.UpdatedAt),
	}
}

// convertWalletToProto converts a models.Wallet to pb.Wallet
func convertWalletToProto(wallet *models.Wallet) *pb.Wallet {
	if wallet == nil {
		return nil
	}

	// Convert wallet accounts
	accounts := make([]*pb.WalletAccount, len(wallet.Accounts))
	for i, account := range wallet.Accounts {
		accounts[i] = convertWalletAccountToProto(&account)
	}

	return &pb.Wallet{
		Id:             wallet.ID.String(),
		OrganizationId: wallet.OrganizationID.String(),
		Name:           wallet.Name,
		PublicKey:      wallet.PublicKey,
		Accounts:       accounts,
		Tags:           wallet.Tags,
		IsActive:       wallet.IsActive,
		CreatedAt:      timestamppb.New(wallet.CreatedAt),
		UpdatedAt:      timestamppb.New(wallet.UpdatedAt),
	}
}

// convertWalletAccountToProto converts a models.WalletAccount to pb.WalletAccount
func convertWalletAccountToProto(account *models.WalletAccount) *pb.WalletAccount {
	if account == nil {
		return nil
	}

	return &pb.WalletAccount{
		Id:            account.ID.String(),
		WalletId:      account.WalletID.String(),
		Name:          account.Name,
		Path:          account.Path,
		PublicKey:     account.PublicKey,
		Address:       account.Address,
		Curve:         account.Curve,
		AddressFormat: account.AddressFormat,
		IsActive:      account.IsActive,
		CreatedAt:     timestamppb.New(account.CreatedAt),
		UpdatedAt:     timestamppb.New(account.UpdatedAt),
	}
}

// convertActivityToProto converts a models.Activity to pb.Activity
func convertActivityToProto(activity *models.Activity) *pb.Activity {
	if activity == nil {
		return nil
	}

	result := ""
	if activity.Result != nil {
		result = string(activity.Result)
	}

	return &pb.Activity{
		Id:             activity.ID.String(),
		OrganizationId: activity.OrganizationID.String(),
		Type:           activity.Type,
		Status:         activity.Status,
		Parameters:     string(activity.Parameters),
		Result:         &result,
		Intent:         convertActivityIntentToProto(&activity.Intent),
		CreatedBy:      activity.CreatedBy.String(),
		CreatedAt:      timestamppb.New(activity.CreatedAt),
		UpdatedAt:      timestamppb.New(activity.UpdatedAt),
	}
}

// convertActivityIntentToProto converts a models.ActivityIntent to pb.ActivityIntent
func convertActivityIntentToProto(intent *models.ActivityIntent) *pb.ActivityIntent {
	if intent == nil {
		return nil
	}

	return &pb.ActivityIntent{
		Fingerprint: intent.Fingerprint,
		Summary:     intent.Summary,
	}
}
