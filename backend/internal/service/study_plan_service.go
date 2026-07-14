package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"course_agent_backend/internal/agent/retrieval"
	agentruntime "course_agent_backend/internal/agent/types"
	apperrors "course_agent_backend/internal/errors"
	"course_agent_backend/internal/model"
	"course_agent_backend/internal/repository"
	"course_agent_backend/internal/vo"
)

type StudyPlanService struct {
	courseRepo    *repository.CourseRepository
	materialRepo  *repository.MaterialRepository
	studyPlanRepo *repository.StudyPlanRepository
	planner       *studyPlanPlanner
}

func NewStudyPlanService(
	courseRepo *repository.CourseRepository,
	materialRepo *repository.MaterialRepository,
	studyPlanRepo *repository.StudyPlanRepository,
	planner *studyPlanPlanner,
) *StudyPlanService {
	return &StudyPlanService{
		courseRepo:    courseRepo,
		materialRepo:  materialRepo,
		studyPlanRepo: studyPlanRepo,
		planner:       planner,
	}
}

func (s *StudyPlanService) ListPlans(ctx context.Context, userID, courseID uint64) ([]vo.StudyPlanSummaryVO, error) {
	if _, err := s.requireActiveCourseMember(ctx, userID, courseID); err != nil {
		return nil, err
	}
	plans, err := s.studyPlanRepo.ListPlansByCourseIDAndUserID(ctx, courseID, userID)
	if err != nil {
		return nil, err
	}
	result := make([]vo.StudyPlanSummaryVO, 0, len(plans))
	for _, plan := range plans {
		items, err := s.studyPlanRepo.ListItemsByPlanID(ctx, plan.ID)
		if err != nil {
			return nil, err
		}
		doneCount := 0
		for _, item := range items {
			if item.Status == "done" {
				doneCount++
			}
		}
		result = append(result, vo.StudyPlanSummaryVO{
			ID:               plan.ID,
			CourseID:         plan.CourseID,
			UserID:           plan.UserID,
			Goal:             plan.Goal,
			DeadlineDate:     plan.DeadlineDate,
			DailyMinutes:     plan.DailyMinutes,
			Status:           plan.Status,
			GeneratedSummary: plan.GeneratedSummary,
			ItemCount:        len(items),
			DoneItemCount:    doneCount,
			CreatedAt:        plan.CreatedAt,
			UpdatedAt:        plan.UpdatedAt,
		})
	}
	return result, nil
}

func (s *StudyPlanService) GetPlan(ctx context.Context, userID, courseID, planID uint64) (*vo.StudyPlanVO, error) {
	if _, err := s.requireActiveCourseMember(ctx, userID, courseID); err != nil {
		return nil, err
	}
	plan, err := s.studyPlanRepo.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if plan.CourseID != courseID || plan.UserID != userID {
		return nil, apperrors.ErrForbidden
	}
	return s.loadPlanVO(ctx, plan)
}

func (s *StudyPlanService) GeneratePlan(ctx context.Context, userID, courseID uint64, goal, deadlineDate string, dailyMinutes int) (*vo.StudyPlanVO, error) {
	if _, err := s.requireActiveCourseMember(ctx, userID, courseID); err != nil {
		return nil, err
	}
	trimmedGoal := strings.TrimSpace(goal)
	if trimmedGoal == "" || dailyMinutes < 15 || dailyMinutes > 480 {
		return nil, apperrors.ErrInvalidParameter
	}
	deadline, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(deadlineDate), time.Local)
	if err != nil {
		return nil, apperrors.ErrInvalidParameter
	}
	today := dateOnly(time.Now())
	if deadline.Before(today) {
		return nil, apperrors.ErrInvalidParameter
	}
	totalDays := int(deadline.Sub(today).Hours()/24) + 1
	if totalDays <= 0 || totalDays > 180 {
		return nil, apperrors.ErrInvalidParameter
	}

	materials, err := s.loadMaterialContexts(ctx, courseID)
	if err != nil {
		return nil, err
	}

	result, planErr := s.planner.Generate(ctx, studyPlanGenerationInput{
		Goal:         trimmedGoal,
		DeadlineDate: deadline,
		DailyMinutes: dailyMinutes,
		Materials:    materials,
	})
	if planErr != nil {
		log.Printf("study plan generation fallback: course_id=%d user_id=%d err=%v", courseID, userID, planErr)
		result = buildFallbackStudyPlan(trimmedGoal, deadline, dailyMinutes, materials)
	}

	planModel := &model.StudyPlan{
		CourseID:         courseID,
		UserID:           userID,
		Goal:             trimmedGoal,
		DeadlineDate:     deadline,
		DailyMinutes:     dailyMinutes,
		Status:           "active",
		GeneratedSummary: strings.TrimSpace(result.Summary),
	}
	itemModels := make([]model.StudyPlanItem, 0, len(result.Items))
	for _, item := range result.Items {
		materialIDsJSON, err := marshalUint64Slice(item.MaterialNodeIDs)
		if err != nil {
			return nil, err
		}
		itemModels = append(itemModels, model.StudyPlanItem{
			DayIndex:         item.DayIndex,
			PlanDate:         item.PlanDate,
			Title:            item.Title,
			TasksText:        item.TasksText,
			SuggestedMinutes: item.SuggestedMinutes,
			MaterialNodeIDs:  materialIDsJSON,
			Status:           "pending",
		})
	}

	err = s.studyPlanRepo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.studyPlanRepo.CreatePlanTx(tx, planModel); err != nil {
			return err
		}
		for index := range itemModels {
			itemModels[index].PlanID = planModel.ID
		}
		return s.studyPlanRepo.CreateItemsTx(tx, itemModels)
	})
	if err != nil {
		return nil, err
	}

	return s.loadPlanVO(ctx, planModel)
}

func (s *StudyPlanService) UpdatePlanItemStatus(ctx context.Context, userID, courseID, planID, itemID uint64, status string) (*vo.StudyPlanVO, error) {
	if _, err := s.requireActiveCourseMember(ctx, userID, courseID); err != nil {
		return nil, err
	}
	if status != "pending" && status != "done" {
		return nil, apperrors.ErrInvalidParameter
	}
	plan, err := s.studyPlanRepo.GetPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if plan.CourseID != courseID || plan.UserID != userID {
		return nil, apperrors.ErrForbidden
	}
	item, err := s.studyPlanRepo.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidParameter
		}
		return nil, err
	}
	if item.PlanID != plan.ID {
		return nil, apperrors.ErrInvalidParameter
	}
	item.Status = status
	if err := s.studyPlanRepo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}
	return s.loadPlanVO(ctx, plan)
}

func (s *StudyPlanService) loadPlanVO(ctx context.Context, plan *model.StudyPlan) (*vo.StudyPlanVO, error) {
	items, err := s.studyPlanRepo.ListItemsByPlanID(ctx, plan.ID)
	if err != nil {
		return nil, err
	}
	itemVOs := make([]vo.StudyPlanItemVO, 0, len(items))
	for _, item := range items {
		materialIDs, err := unmarshalUint64Slice(item.MaterialNodeIDs)
		if err != nil {
			return nil, err
		}
		itemVOs = append(itemVOs, vo.StudyPlanItemVO{
			ID:               item.ID,
			PlanID:           item.PlanID,
			DayIndex:         item.DayIndex,
			PlanDate:         item.PlanDate,
			Title:            item.Title,
			TasksText:        item.TasksText,
			SuggestedMinutes: item.SuggestedMinutes,
			MaterialNodeIDs:  materialIDs,
			Status:           item.Status,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		})
	}
	result := &vo.StudyPlanVO{
		ID:               plan.ID,
		CourseID:         plan.CourseID,
		UserID:           plan.UserID,
		Goal:             plan.Goal,
		DeadlineDate:     plan.DeadlineDate,
		DailyMinutes:     plan.DailyMinutes,
		Status:           plan.Status,
		GeneratedSummary: plan.GeneratedSummary,
		CreatedAt:        plan.CreatedAt,
		UpdatedAt:        plan.UpdatedAt,
		Items:            itemVOs,
	}
	return result, nil
}

func (s *StudyPlanService) requireActiveCourseMember(ctx context.Context, userID, courseID uint64) (*model.CourseMember, error) {
	course, err := s.courseRepo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrCourseNotFound
		}
		return nil, err
	}
	if course.Status != "active" {
		return nil, apperrors.ErrCourseNotFound
	}
	member, err := s.courseRepo.GetMember(ctx, courseID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrForbidden
		}
		return nil, err
	}
	if member.JoinStatus != "active" {
		return nil, apperrors.ErrForbidden
	}
	return member, nil
}

func (s *StudyPlanService) loadMaterialContexts(ctx context.Context, courseID uint64) ([]studyPlanMaterialContext, error) {
	nodes, err := s.materialRepo.ListActiveFileNodesByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	contexts := make([]studyPlanMaterialContext, 0, len(nodes))
	for _, node := range nodes {
		if strings.TrimSpace(node.StoragePath) == "" {
			continue
		}
		excerpt, readErr := retrieval.ReadMaterialExcerpt(agentruntime.MaterialDocument{
			MaterialNodeID: node.ID,
			FileName:       node.NodeName,
			StoragePath:    node.StoragePath,
			MimeType:       node.MimeType,
		}, 800)
		if readErr != nil {
			excerpt = ""
		}
		contexts = append(contexts, studyPlanMaterialContext{
			MaterialNodeID: node.ID,
			FileName:       node.NodeName,
			Summary:        strings.TrimSpace(excerpt),
		})
		if len(contexts) >= 8 {
			break
		}
	}
	return contexts, nil
}

type studyPlanPlanner struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewStudyPlanPlanner(baseURL, apiKey, model string) *studyPlanPlanner {
	return &studyPlanPlanner{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		apiKey:  strings.TrimSpace(apiKey),
		model:   strings.TrimSpace(model),
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

type studyPlanGenerationInput struct {
	Goal         string
	DeadlineDate time.Time
	DailyMinutes int
	Materials    []studyPlanMaterialContext
}

type studyPlanMaterialContext struct {
	MaterialNodeID uint64
	FileName       string
	Summary        string
}

type studyPlanGenerationResult struct {
	Summary string
	Items   []studyPlanGeneratedItem
}

type studyPlanGeneratedItem struct {
	DayIndex         int
	PlanDate         time.Time
	Title            string
	TasksText        string
	SuggestedMinutes int
	MaterialNodeIDs  []uint64
}

func (p *studyPlanPlanner) Generate(ctx context.Context, input studyPlanGenerationInput) (studyPlanGenerationResult, error) {
	if p.baseURL == "" || p.apiKey == "" || p.model == "" {
		return studyPlanGenerationResult{}, fmt.Errorf("planner config unavailable")
	}
	responseBody, err := p.requestPlan(ctx, input)
	if err != nil {
		return studyPlanGenerationResult{}, err
	}
	return parseStudyPlanResult(responseBody, input)
}

func (p *studyPlanPlanner) requestPlan(ctx context.Context, input studyPlanGenerationInput) (string, error) {
	systemPrompt := "你是课程学习计划助手。你只能基于用户目标和当前课程资料制定学习计划。请输出 JSON 对象，格式为 {\"summary\":\"...\",\"items\":[{\"dayIndex\":1,\"title\":\"...\",\"tasksText\":\"...\",\"suggestedMinutes\":60,\"materialNodeIds\":[1,2]}]}。不要输出 Markdown。不要输出代码块。"
	userPrompt := buildStudyPlanPrompt(input)
	payload := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.4,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("planner http %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("planner returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func buildStudyPlanPrompt(input studyPlanGenerationInput) string {
	var builder strings.Builder
	builder.WriteString("学习目标：")
	builder.WriteString(input.Goal)
	builder.WriteString("\n截止日期：")
	builder.WriteString(input.DeadlineDate.Format("2006-01-02"))
	builder.WriteString("\n每日可用时间（分钟）：")
	builder.WriteString(fmt.Sprintf("%d", input.DailyMinutes))
	builder.WriteString("\n请从今天开始按天安排任务。")
	if len(input.Materials) > 0 {
		builder.WriteString("\n当前课程资料：")
		for _, material := range input.Materials {
			builder.WriteString("\n- [")
			builder.WriteString(fmt.Sprintf("%d", material.MaterialNodeID))
			builder.WriteString("] ")
			builder.WriteString(material.FileName)
			if strings.TrimSpace(material.Summary) != "" {
				builder.WriteString("：")
				builder.WriteString(material.Summary)
			}
		}
	}
	return builder.String()
}

func parseStudyPlanResult(raw string, input studyPlanGenerationInput) (studyPlanGenerationResult, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return studyPlanGenerationResult{}, fmt.Errorf("planner output is not json")
	}
	cleaned := raw[start : end+1]
	var parsed struct {
		Summary string `json:"summary"`
		Items   []struct {
			DayIndex         int      `json:"dayIndex"`
			Title            string   `json:"title"`
			TasksText        string   `json:"tasksText"`
			SuggestedMinutes int      `json:"suggestedMinutes"`
			MaterialNodeIDs  []uint64 `json:"materialNodeIds"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return studyPlanGenerationResult{}, err
	}
	if len(parsed.Items) == 0 {
		return studyPlanGenerationResult{}, fmt.Errorf("planner output has no items")
	}
	today := dateOnly(time.Now())
	maxDays := int(input.DeadlineDate.Sub(today).Hours()/24) + 1
	result := studyPlanGenerationResult{Summary: strings.TrimSpace(parsed.Summary)}
	for index, item := range parsed.Items {
		dayIndex := item.DayIndex
		if dayIndex <= 0 {
			dayIndex = index + 1
		}
		if dayIndex > maxDays {
			break
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("第 %d 天学习任务", dayIndex)
		}
		tasksText := strings.TrimSpace(item.TasksText)
		if tasksText == "" {
			tasksText = "完成当天学习任务，并记录重点内容。"
		}
		suggestedMinutes := item.SuggestedMinutes
		if suggestedMinutes <= 0 || suggestedMinutes > input.DailyMinutes {
			suggestedMinutes = input.DailyMinutes
		}
		result.Items = append(result.Items, studyPlanGeneratedItem{
			DayIndex:         dayIndex,
			PlanDate:         today.AddDate(0, 0, dayIndex-1),
			Title:            title,
			TasksText:        tasksText,
			SuggestedMinutes: suggestedMinutes,
			MaterialNodeIDs:  item.MaterialNodeIDs,
		})
	}
	if len(result.Items) == 0 {
		return studyPlanGenerationResult{}, fmt.Errorf("planner output has no valid items")
	}
	return normalizeStudyPlanResult(result, input), nil
}

func buildFallbackStudyPlan(goal string, deadline time.Time, dailyMinutes int, materials []studyPlanMaterialContext) studyPlanGenerationResult {
	totalDays := int(deadline.Sub(dateOnly(time.Now())).Hours()/24) + 1
	result := studyPlanGenerationResult{
		Summary: "已根据学习目标、截止时间和每日可用时间生成学习计划。",
		Items:   make([]studyPlanGeneratedItem, 0, totalDays),
	}
	for dayIndex := 1; dayIndex <= totalDays; dayIndex++ {
		var title string
		switch {
		case dayIndex == 1:
			title = "梳理目标与资料"
		case dayIndex == totalDays:
			title = "集中复盘与查漏补缺"
		case dayIndex*100/totalDays <= 60:
			title = "推进核心内容学习"
		case dayIndex*100/totalDays <= 85:
			title = "巩固理解与整理笔记"
		default:
			title = "复习重点与自测"
		}
		referenced := selectMaterialIDsForDay(materials, dayIndex)
		taskLines := []string{fmt.Sprintf("围绕“%s”完成当天学习。", goal)}
		if len(referenced) > 0 {
			taskLines = append(taskLines, "优先阅读并整理当天关联资料。")
		}
		if dayIndex == totalDays {
			taskLines = append(taskLines, "回顾前几天笔记，总结薄弱点并完成一次整体复盘。")
		} else {
			taskLines = append(taskLines, "记录关键概念、疑问点和需要复习的内容。")
		}
		result.Items = append(result.Items, studyPlanGeneratedItem{
			DayIndex:         dayIndex,
			PlanDate:         dateOnly(time.Now()).AddDate(0, 0, dayIndex-1),
			Title:            title,
			TasksText:        strings.Join(taskLines, " "),
			SuggestedMinutes: dailyMinutes,
			MaterialNodeIDs:  referenced,
		})
	}
	return normalizeStudyPlanResult(result, studyPlanGenerationInput{
		Goal:         goal,
		DeadlineDate: deadline,
		DailyMinutes: dailyMinutes,
		Materials:    materials,
	})
}

func normalizeStudyPlanResult(result studyPlanGenerationResult, input studyPlanGenerationInput) studyPlanGenerationResult {
	today := dateOnly(time.Now())
	maxDays := int(input.DeadlineDate.Sub(today).Hours()/24) + 1
	if maxDays <= 0 {
		maxDays = len(result.Items)
	}
	normalized := studyPlanGenerationResult{
		Summary: strings.TrimSpace(result.Summary),
		Items:   make([]studyPlanGeneratedItem, 0, maxDays),
	}
	if normalized.Summary == "" {
		normalized.Summary = "已生成学习计划。"
	}
	for index, item := range result.Items {
		if index >= maxDays {
			break
		}
		dayIndex := index + 1
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("第 %d 天学习任务", dayIndex)
		}
		tasksText := strings.TrimSpace(item.TasksText)
		if tasksText == "" {
			tasksText = "完成当天学习内容，并记录关键收获。"
		}
		suggestedMinutes := item.SuggestedMinutes
		if suggestedMinutes <= 0 || suggestedMinutes > input.DailyMinutes {
			suggestedMinutes = input.DailyMinutes
		}
		normalized.Items = append(normalized.Items, studyPlanGeneratedItem{
			DayIndex:         dayIndex,
			PlanDate:         today.AddDate(0, 0, dayIndex-1),
			Title:            title,
			TasksText:        tasksText,
			SuggestedMinutes: suggestedMinutes,
			MaterialNodeIDs:  uniqueMaterialIDs(item.MaterialNodeIDs),
		})
	}
	for len(normalized.Items) < maxDays {
		dayIndex := len(normalized.Items) + 1
		normalized.Items = append(normalized.Items, studyPlanGeneratedItem{
			DayIndex:         dayIndex,
			PlanDate:         today.AddDate(0, 0, dayIndex-1),
			Title:            fmt.Sprintf("第 %d 天学习任务", dayIndex),
			TasksText:        "根据当前学习目标继续推进学习，并整理当天重点与疑问。",
			SuggestedMinutes: input.DailyMinutes,
			MaterialNodeIDs:  selectMaterialIDsForDay(input.Materials, dayIndex),
		})
	}
	return normalized
}

func selectMaterialIDsForDay(materials []studyPlanMaterialContext, dayIndex int) []uint64 {
	if len(materials) == 0 {
		return nil
	}
	first := materials[(dayIndex-1)%len(materials)].MaterialNodeID
	result := []uint64{first}
	if len(materials) > 1 && dayIndex%2 == 0 {
		second := materials[dayIndex%len(materials)].MaterialNodeID
		if second != first {
			result = append(result, second)
		}
	}
	return result
}

func uniqueMaterialIDs(ids []uint64) []uint64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[uint64]struct{}, len(ids))
	result := make([]uint64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func marshalUint64Slice(values []uint64) (string, error) {
	if len(values) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(uniqueMaterialIDs(values))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalUint64Slice(raw string) ([]uint64, error) {
	if strings.TrimSpace(raw) == "" {
		return []uint64{}, nil
	}
	var values []uint64
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
