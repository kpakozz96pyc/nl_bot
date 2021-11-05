package database

import (
	"main/models"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type ConcurrentNeighborRepository struct {
	db sqlx.DB
}

func NewConcurrentNeighborRepository(database sqlx.DB) *ConcurrentNeighborRepository {
	database.SetMaxIdleConns(20)
	database.SetMaxOpenConns(200)
	return &ConcurrentNeighborRepository{db: database}
}

func (cnr ConcurrentNeighborRepository) GetAll() ([]models.Neighbor, error) {

	var neighbors []models.Neighbor
	err := cnr.db.Select(&neighbors, "select * from neighbors")
	if err != nil {
		return nil, err
	}

	return neighbors, nil
}

func (cnr ConcurrentNeighborRepository) Insert(n models.Neighbor) error {

	query := `INSERT INTO neighbors(name, telegram_first_name, telegram_last_name, telegram_user_name, section, building, flat)
							 VALUES(:name,:telegram_first_name,:telegram_last_name, :telegram_user_name, :section, :building, :flat)`

	_, er := cnr.db.NamedExec(query, &n)
	if er != nil {
		return er
	}

	return nil
}

func (cnr ConcurrentNeighborRepository) Upsert(n models.Neighbor) error {
	dbRecords, err := cnr.GetByTelegramName(n.TelegramUserName)
	if err != nil {
		return err
	}
	if len(dbRecords) > 0 {
		return cnr.Update(n)
	} else {
		return cnr.Insert(n)
	}
}

func (cnr ConcurrentNeighborRepository) Update(n models.Neighbor) error {

	query := `Update neighbors 
	set name = :name, 
	telegram_first_name = :telegram_first_name, 
	telegram_last_name = :telegram_last_name, 
	section = :section, 
	building = :building, 
	flat = :flat
	where telegram_user_name = :telegram_user_name`

	_, er := cnr.db.NamedExec(query, &n)
	if er != nil {
		return er
	}

	return nil
}

func (cnr ConcurrentNeighborRepository) Delete(telegramUserName string) error {

	query := `Delete from neighbors where telegram_user_name =$1`

	_, er := cnr.db.Exec(query, telegramUserName)
	if er != nil {
		return er
	}

	return nil

}

func (cnr ConcurrentNeighborRepository) GetByTelegramName(telegramUserName string) ([]models.Neighbor, error) {

	var neighbors []models.Neighbor
	err := cnr.db.Select(&neighbors, "select * from neighbors where telegram_user_name = '"+telegramUserName+"'")
	if err != nil {
		return nil, err
	}

	return neighbors, nil
}
