package service

import (
	"context"
	"errors"
	"strings"
)

type distributionUserChannelMemberRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]DistributionMemberView, error)
}

type distributionUserChannelOrganizationRepository interface {
	GetByOwnerUserID(ctx context.Context, userID int64) (*DistributionOrganization, error)
}

type distributionUserChannelAttributionRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*DistributionAttribution, error)
}

func resolveDistributionUserChannelOrgID(
	ctx context.Context,
	userID int64,
	memberRepo distributionUserChannelMemberRepository,
	organizationRepo distributionUserChannelOrganizationRepository,
	attributionRepo distributionUserChannelAttributionRepository,
) (int64, error) {
	if userID <= 0 {
		return 0, ErrInvalidDistributionAttribution
	}

	if memberRepo != nil {
		members, err := memberRepo.ListByUserID(ctx, userID)
		if err != nil {
			return 0, err
		}
		channelOrgID := int64(0)
		for _, member := range members {
			if member.ChannelOrgID <= 0 {
				continue
			}
			if channelOrgID == 0 {
				channelOrgID = member.ChannelOrgID
				continue
			}
			if channelOrgID != member.ChannelOrgID {
				return 0, ErrDistributionMemberChannelConflict
			}
		}
		if channelOrgID > 0 {
			return channelOrgID, nil
		}
	}

	if organizationRepo != nil {
		org, err := organizationRepo.GetByOwnerUserID(ctx, userID)
		if err != nil {
			return 0, err
		}
		if org != nil && org.ID > 0 && strings.EqualFold(strings.TrimSpace(org.Status), "active") {
			return org.ID, nil
		}
	}

	if attributionRepo != nil {
		attribution, err := attributionRepo.GetByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, ErrDistributionAttributionNotFound) {
				return 0, err
			}
			return 0, err
		}
		if attribution != nil && attribution.ChannelOrgID > 0 {
			return attribution.ChannelOrgID, nil
		}
	}

	return 0, ErrDistributionAttributionNotFound
}
