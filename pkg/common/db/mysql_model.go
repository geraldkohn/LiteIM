package db

import "gorm.io/gorm"

func create(db *gorm.DB, item interface{}) error {
	err := db.Create(item).Error
	if err != nil {
		// logs.Error(result.Error)
		return err
	}
	return nil
}

func createBatch(db *gorm.DB, list interface{}) error {
	err := db.Create(list).Error
	if err != nil {
		// logs.Error(result.Error)
		return err
	}
	return nil
}

func update(db *gorm.DB, item interface{}, data map[string]interface{}) error {
	err := db.Model(item).Updates(data).Error
	if err != nil {
		// logs.Error(result.Error)
		return err
	}
	return nil
}

func getObjectByID(db *gorm.DB, item interface{}, id int) error {
	err := db.First(item, id).Error
	if err != nil {
		// logs.Error(result.Error)
		return err
	}
	return nil
}

func getObjectByCond(db *gorm.DB, item interface{}, order string, conds map[string]interface{}) error {
	tx := db
	for cond, value := range conds {
		tx = tx.Where(cond, value)
	}
	if order != "" {
		tx = tx.Order(order)
	}
	err := tx.First(item).Error
	if err != nil {
		// // logs.Error(result.Error)
		return err
	}
	return nil
}

func getListBatch(db *gorm.DB, item interface{}, order string, page, limit int, conds map[string]interface{}) error {

	//tx := db.Where("1 = 1")
	tx := db
	for cond, value := range conds {
		tx = tx.Where(cond, value)
	}
	if order != "" {
		tx = tx.Order(order)
	}
	if page >= 0 && limit > 0 {
		tx = tx.Limit(limit).Offset(limit * (page - 1))
	}
	err := tx.Find(item).Error
	if err != nil {
		// // logs.Error(result.Error)
		return err
	}
	return nil
}

func getTotal(db *gorm.DB, tableName string, conds map[string]interface{}) int64 {
	var total int64

	//	tx := db.Where("1 = 1")
	tx := db
	tx = tx.Table(tableName)
	for cond, value := range conds {
		tx = tx.Where(cond, value)
	}
	err := tx.Count(&total).Error
	if err != nil {
		// // logs.Error(result.Error)
		return 0
	}
	return total
}
