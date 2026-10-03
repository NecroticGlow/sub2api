package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type knowledgeQualityPlanRepo struct {
	ScheduledTestPlanRepository
	created int
	updated int
}

func (r *knowledgeQualityPlanRepo) Create(_ context.Context, plan *ScheduledTestPlan) (*ScheduledTestPlan, error) {
	r.created++
	return plan, nil
}

func (r *knowledgeQualityPlanRepo) Update(_ context.Context, plan *ScheduledTestPlan) (*ScheduledTestPlan, error) {
	r.updated++
	return plan, nil
}

func knowledgeQualityPlan(kind, action string) *ScheduledTestPlan {
	plan := pelicanPlan()
	plan.PelicanConfig.QuestionKind = kind
	plan.PelicanConfig.TestChannel = "account"
	plan.PelicanConfig.Prompt = "仅依据已有知识，不联网、不调用工具、不猜测。"
	answer := "iPhone 17 series; September 9, 2025; September 19, 2025; GeForce RTX 5090; Android 16; macOS Tahoe 26; Windows 11, version 25H2"
	if kind == "japan_pm" {
		answer = "高市早苗"
	}
	plan.PelicanConfig.Quality = &QualityPolicy{
		ExpectedAnswer: answer, Action: action, RemoveGroupIDs: []int64{3},
		Judge: &QualityJudgeConfig{GroupID: 9, ModelID: "mock-judge", Prompt: "Compare reference_answer and candidate_answer."},
	}
	return plan
}

func TestQualityKnowledgePlanCreateUpdateAndTemplate(t *testing.T) {
	for _, kind := range []string{"knowledge", "japan_pm"} {
		for _, action := range []string{"observe_only", "remove_groups", "disable_scheduling"} {
			t.Run(kind+"/"+action, func(t *testing.T) {
				plan := knowledgeQualityPlan(kind, action)
				repo := &knowledgeQualityPlanRepo{}
				svc := NewScheduledTestService(repo, nil)
				created, err := svc.CreatePlan(context.Background(), plan)
				require.NoError(t, err)
				require.NotNil(t, created.NextRunAt)
				updated, err := svc.UpdatePlan(context.Background(), plan)
				require.NoError(t, err)
				require.Equal(t, 1, repo.created)
				require.Equal(t, 1, repo.updated)
				require.Equal(t, kind, updated.PelicanConfig.QuestionKind)
				require.NotEmpty(t, updated.PelicanConfig.Quality.ExpectedAnswer)
				require.NotContains(t, intelligenceTestPrompt(updated.PelicanConfig), "HTML")
				template := &QualityRuleTemplate{ModelID: plan.ModelID, CronExpression: plan.CronExpression,
					Enabled: true, MaxResults: 50, PelicanConfig: cloneQualityPelicanConfig(plan.PelicanConfig)}
				require.NoError(t, prepareQualityTemplate(template), "group rules must accept both knowledge questions too")
			})
		}
	}
}

func TestQualityKnowledgePlanPreservesValidationGuards(t *testing.T) {
	changes := map[string]func(*ScheduledTestPlan){
		"HTML question":    func(p *ScheduledTestPlan) { p.PelicanConfig.QuestionKind = "pelican" },
		"unknown question": func(p *ScheduledTestPlan) { p.PelicanConfig.QuestionKind = "unknown" },
		"empty reference":  func(p *ScheduledTestPlan) { p.PelicanConfig.Quality.ExpectedAnswer = " " },
		"invalid judge":    func(p *ScheduledTestPlan) { p.PelicanConfig.Quality.Judge.GroupID = 0 },
		"BPS channel":      func(p *ScheduledTestPlan) { p.PelicanConfig.TestChannel = "bps" },
		"BPS action":       func(p *ScheduledTestPlan) { p.PelicanConfig.Quality.Action = QualityActionEnableBPS },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			plan := knowledgeQualityPlan("knowledge", "observe_only")
			change(plan)
			repo := &knowledgeQualityPlanRepo{}
			svc := NewScheduledTestService(repo, nil)
			_, err := svc.CreatePlan(context.Background(), plan)
			require.Error(t, err)
			_, err = svc.UpdatePlan(context.Background(), plan)
			require.Error(t, err)
			require.Zero(t, repo.created)
			require.Zero(t, repo.updated)
		})
	}
}
