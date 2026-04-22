package handler

import "github.com/gladinov/effective_mobile_test_assignment/internal/domain"

func mapYearMonthPtrToDomain(y *yearMonth) *domain.YearMonth {
	if y == nil {
		return nil
	}
	return &domain.YearMonth{
		Year:  y.Year,
		Month: y.Month,
	}
}

func mapDomainSubToDTOSubResponce(domainSub domain.Subscription) subscriptionResponce {
	return subscriptionResponce{
		ID:          domainSub.ID,
		ServiceName: domainSub.ServiceName,
		Price:       domainSub.Price,
		UserID:      domainSub.UserID,
		StartDate:   MapDomainYearMonthToDto(domainSub.StartDate),
		EndDate:     MapDomainYearMonthPtrToDtoPtr(domainSub.EndDate),
	}
}

func MapDomainYearMonthToDto(date domain.YearMonth) yearMonth {
	return yearMonth{
		Year:  date.Year,
		Month: date.Month,
	}
}

func MapDomainYearMonthPtrToDtoPtr(date *domain.YearMonth) *yearMonth {
	if date == nil {
		return nil
	}
	return &yearMonth{
		Year:  date.Year,
		Month: date.Month,
	}
}

func mapYearMonthPtrToDomainPtr(date *yearMonth) *domain.YearMonth {
	if date == nil {
		return nil
	}
	return &domain.YearMonth{
		Year:  date.Year,
		Month: date.Month,
	}
}
