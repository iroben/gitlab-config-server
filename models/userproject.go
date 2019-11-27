package models

type UserProject struct {
	BaseModel
	Id        int `gorm:"column(id)" json:"id"`
	UserId    int `gorm:"column(user_id)" jjson:"user_id"`
	ProjectId int `gorm:"column(project_id)" jjson:"project_id"`
}

func (UserProject) TableName() string {
	return "user_project"
}
func (m UserProject) UserProjects() []*GitlabProject {
	var userProjects []UserProject
	err := DB.Where("user_id=?", m.UserId).Find(&userProjects).Error
	if err != nil {
		(&m).Error(err)
		return nil
	}
	projectIds := make([]int, len(userProjects))
	for index, userProject := range userProjects {
		projectIds[index] = userProject.ProjectId
	}
	return GitlabProject{}.FindByIds(projectIds)

}
func (m UserProject) SaveProjects(projects []*GitlabProject) bool {
	tx := DB.Begin()
	err := tx.Where("user_id=?", m.UserId).Delete(UserProject{}).Error
	if err != nil {
		m.Error(err)
		tx.Rollback()
		return false
	}
	for _, project := range projects {
		err := tx.Create(&UserProject{
			UserId:    m.UserId,
			ProjectId: project.Id,
		}).Error
		if err != nil {
			m.Error(err)
			tx.Rollback()
			return false
		}
	}
	tx.Commit()
	return true
}
