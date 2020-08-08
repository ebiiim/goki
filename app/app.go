package app

import (
	"fmt"
	"time"

	"github.com/ebiiim/goki"
	"github.com/ebiiim/goki/db"
	"github.com/ebiiim/goki/model"
)

type App struct {
	Users      db.UserDB
	Activities db.ActivityDB
}

func NewApp(userDB db.UserDB, activityDB db.ActivityDB) *App {
	a := &App{
		Users:      userDB,
		Activities: activityDB,
	}
	return a
}

func (a *App) Close() error {
	err1 := a.Activities.Close()
	err2 := a.Users.Close()
	if err1 == nil && err2 == nil {
		return nil
	}
	return goki.ErrAppClose(fmt.Errorf("close UserDB=%v ActivityDB=%v", err2, err1))
}

func (a *App) AddUser(userID, userName, twitterID string) (*model.User, error) {
	u := model.NewUser(userID, userName, twitterID)
	if err := a.Users.Add(u); err != nil {
		return nil, fmt.Errorf("App.AddUser: %w", err)
	}
	return u, nil
}

func (a *App) GetUser(userID string) (*model.User, error) {
	u, err := a.Users.Get(userID)
	if err != nil {
		return nil, fmt.Errorf("App.GetUser: %w", err)
	}
	return u, nil
}

func (a *App) Action(userID string, numS, numM, numL int) (*model.Activity, error) {
	act := model.NewActivity(userID, goki.TimeNow(), numS, numM, numL)
	if err := a.Activities.Add(act); err != nil {
		return nil, fmt.Errorf("App.Action: %w", err)
	}
	return act, nil
}

func (a *App) CountByYear(userID string, year int, tz ...*time.Location) (*model.Goki, error) {
	loc := time.UTC
	if len(tz) != 0 {
		loc = tz[0]
	}
	begin := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	year++
	end := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	return a.count(userID, begin, end)
}

func (a *App) CountByMonth(userID string, year int, month time.Month, tz ...*time.Location) (*model.Goki, error) {
	loc := time.UTC
	if len(tz) != 0 {
		loc = tz[0]
	}
	begin := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	endMonth := month + 1
	if month == time.December {
		endMonth = time.January
	}
	end := time.Date(year, endMonth, 1, 0, 0, 0, 0, loc)
	return a.count(userID, begin, end)
}

func (a *App) count(userID string, begin, end time.Time) (*model.Goki, error) {
	filter := db.QueryFuncTime(begin, end)
	acts, err := a.Activities.Query(userID, filter)
	if err != nil {
		return nil, fmt.Errorf("App.CountBy*: %w", err)
	}
	gs := make([]*model.Goki, len(acts))
	for idx, v := range acts {
		gs[idx] = v.G
	}
	return model.GokiSum(gs...), nil
}
