package services

import (
	"context"
	"log/slog"

	"github.com/jm96441n/movieswithfriends/identityaccess"
	iamstore "github.com/jm96441n/movieswithfriends/identityaccess/store"
	"github.com/jm96441n/movieswithfriends/metrics"
	"github.com/jm96441n/movieswithfriends/partymgmt"
	partymgmtstore "github.com/jm96441n/movieswithfriends/partymgmt/store"
)

type ProfileAggregatorService struct {
	profileRepository *iamstore.ProfileRepository
	watcherRepository *partymgmtstore.WatcherRepository
}

func NewProfileAggregatorService(profileRepository *iamstore.ProfileRepository, watcherRepository *partymgmtstore.WatcherRepository) *ProfileAggregatorService {
	return &ProfileAggregatorService{
		profileRepository: profileRepository,
		watcherRepository: watcherRepository,
	}
}

type MovieData struct {
	NumPages      int
	CurPage       int
	WatchedMovies []partymgmt.PartyMovie
}

type ProfilePageData struct {
	Profile *identityaccess.Profile
	Parties []partymgmt.Party

	NumPages      int
	CurPage       int
	WatchedMovies []partymgmt.PartyMovie
}

type profileResult struct {
	profile *identityaccess.Profile
	err     error
}

type statsResult struct {
	stats identityaccess.ProfileStats
	err   error
}

type partiesResult struct {
	parties []partymgmt.Party
	err     error
}

type watchHistoryResult struct {
	movieData MovieData
	err       error
}

func (p *ProfileAggregatorService) GetProfilePageData(ctx context.Context, logger *slog.Logger, profileID int) (ProfilePageData, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "profileAggregatorService.GetProfilePageData")
	defer span.End()
	profResultCh := make(chan profileResult)
	statsResultCh := make(chan statsResult)
	partiesResultCh := make(chan partiesResult)
	watchHistoryResultCh := make(chan watchHistoryResult)

	go func() {
		profile, err := p.getProfile(ctx, profileID)
		profResultCh <- profileResult{profile: profile, err: err}
	}()

	go func() {
		stats, err := p.getProfileStats(ctx, logger, profileID)
		statsResultCh <- statsResult{stats: stats, err: err}
	}()

	go func() {
		parties, err := p.getParties(ctx, profileID)
		partiesResultCh <- partiesResult{parties: parties, err: err}
	}()

	go func() {
		movieData, err := p.GetWatchPaginatedHistory(ctx, logger, profileID, PageInfo{PageNum: 1})
		watchHistoryResultCh <- watchHistoryResult{movieData: movieData, err: err}
	}()

	profRes := <-profResultCh
	statsRes := <-statsResultCh
	partiesRes := <-partiesResultCh
	movieDataRes := <-watchHistoryResultCh

	if profRes.err != nil {
		return ProfilePageData{}, profRes.err
	}

	if statsRes.err != nil {
		return ProfilePageData{}, statsRes.err
	}

	if partiesRes.err != nil {
		return ProfilePageData{}, partiesRes.err
	}

	if movieDataRes.err != nil {
		return ProfilePageData{}, movieDataRes.err
	}

	profRes.profile.Stats = statsRes.stats

	return ProfilePageData{
		Profile:       profRes.profile,
		Parties:       partiesRes.parties,
		NumPages:      movieDataRes.movieData.NumPages,
		CurPage:       movieDataRes.movieData.CurPage,
		WatchedMovies: movieDataRes.movieData.WatchedMovies,
	}, nil
}

func (p *ProfileAggregatorService) getProfile(ctx context.Context, profileID int) (*identityaccess.Profile, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "profileAggregatorService.getProfile")
	defer span.End()
	getProfResult, err := p.profileRepository.GetProfileByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	return &identityaccess.Profile{
		ID:        profileID,
		FirstName: getProfResult.FirstName,
		LastName:  getProfResult.LastName,
		CreatedAt: getProfResult.CreatedAt,
		Account: identityaccess.Account{
			ID:    getProfResult.AccountID,
			Email: getProfResult.AccountEmail,
		},
	}, nil
}

func (p *ProfileAggregatorService) getProfileStats(ctx context.Context, logger *slog.Logger, profileID int) (identityaccess.ProfileStats, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "profileAggregatorService.getProfileStats")
	defer span.End()
	stats, err := p.profileRepository.GetProfileStats(ctx, logger, profileID)
	if err != nil {
		return identityaccess.ProfileStats{}, err
	}

	return identityaccess.ProfileStats{
		NumberOfParties: stats.NumParties,
		WatchTime:       stats.WatchTime,
		MoviesWatched:   stats.MoviesWatched,
	}, err
}

func (p *ProfileAggregatorService) getParties(ctx context.Context, profileID int) ([]partymgmt.Party, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "profileAggregatorService.getParties")
	defer span.End()
	partyRes, err := p.watcherRepository.GetPartiesForWatcher(ctx, profileID, 50)
	if err != nil {
		return nil, err
	}

	parties := make([]partymgmt.Party, 0, len(partyRes))
	for _, party := range partyRes {
		parties = append(parties, partymgmt.Party{
			ID:          party.ID,
			Name:        party.Name,
			MemberCount: party.MemberCount,
			MovieCount:  party.MovieCount,
		})
	}

	return parties, nil
}

type PageInfo struct {
	PageNum  int
	PageSize int
}

const defaultPageSize = 5

func (p *ProfileAggregatorService) GetWatchPaginatedHistory(ctx context.Context, logger *slog.Logger, profileID int, pageInfo PageInfo) (MovieData, error) {
	ctx, span, _ := metrics.SpanFromContext(ctx, "profileAggregatorService.GetWatchPaginatedHistory")
	defer span.End()
	pageNum := pageInfo.PageNum
	pageSize := max(defaultPageSize, pageInfo.PageSize)

	offset := pageSize * (pageNum - 1)

	watchedMovies, err := p.watcherRepository.GetWatchedMoviesForWatcher(ctx, profileID, offset)
	if err != nil {
		return MovieData{}, err
	}

	numMovies, err := p.watcherRepository.GetWatchedMoviesCountForMember(ctx, logger, profileID)
	if err != nil {
		return MovieData{}, err
	}

	numPages := numMovies / pageSize
	if numMovies > numPages*pageSize {
		numPages++
	}

	movies := make([]partymgmt.PartyMovie, 0, len(watchedMovies))

	for _, movie := range watchedMovies {
		movies = append(movies, partymgmt.PartyMovie{
			ID:        movie.IDMovie,
			Title:     movie.Title,
			WatchDate: movie.WatchDate,
			PartyName: movie.PartyName,
		})
	}

	return MovieData{
		NumPages:      numPages,
		CurPage:       pageNum,
		WatchedMovies: movies,
	}, nil
}
