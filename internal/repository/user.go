package repository

import (
	"context"
	"effectiveMobileTestProblem/internal/entity"
	"effectiveMobileTestProblem/internal/model"
	"errors"
	"fmt"
	"strings"
)

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) (string, error) {
	tx, err := r.conn.Begin()
	if err != nil {
		return "", err
	}

	defer tx.Rollback()
	query := `INSERT INTO users (passport_number, passport_series, name, surname, address) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var id string
	passportData := strings.Split(user.PassportSeriesAndNumber, " ")
	passportSeries := passportData[0]
	passportNumber := passportData[1]
	err = r.conn.GetContext(ctx, &id, query, passportNumber, passportSeries, user.Name, user.Surname, user.Address)
	if err != nil {
		return "", err
	}
	if tx.Commit() != nil {
		return "", err
	}
	return id, nil
}

func (r *UserRepository) GetUserById(ctx context.Context, id string) (*entity.UserDB, error) {
	query := `SELECT id, passport_number, passport_series, name, surname, address FROM users WHERE id=$1`
	user := make([]*entity.UserDB, 0, 1)
	err := r.conn.SelectContext(ctx, &user, query, id)
	if err != nil {
		return &entity.UserDB{}, err
	}

	if len(user) == 0 {
		return &entity.UserDB{}, model.ErrNotFound

	}
	return user[0], nil
}

func (r *UserRepository) GetUserByPassport(ctx context.Context, passport string) (*entity.UserDB, error) {
	query := `SELECT id, passport_number, passport_series, name, surname, address FROM users WHERE passport_number=$1 AND passport_series=$2`
	user := make([]*entity.UserDB, 0, 1)
	passportData := strings.Split(passport, " ")
	passportSeries := passportData[0]
	passportNumber := passportData[1]

	err := r.conn.SelectContext(ctx, &user, query, passportNumber, passportSeries)

	if err != nil {
		return &entity.UserDB{}, err
	}
	if len(user) == 0 {
		return &entity.UserDB{}, model.ErrNotFound
	}
	return user[0], nil
}

func (r *UserRepository) GetUsers(ctx context.Context, filterAndPagination *model.FilterAndPagination) ([]*entity.UserDB, error) {
	var where []string
	var args []interface{}

	if v := filterAndPagination.PassportSeries; v != "" {
		where = append(where, "passport_series = $"+fmt.Sprint(len(args)+1))
		args = append(args, v)
	}

	if v := filterAndPagination.PassportNumber; v != "" {
		where = append(where, "passport_number = $"+fmt.Sprint(len(args)+1))
		args = append(args, v)
	}

	if v := filterAndPagination.Name; v != "" {
		where = append(where, "name = $"+fmt.Sprint(len(args)+1))
		args = append(args, v)
	}

	if v := filterAndPagination.Surname; v != "" {
		where = append(where, "surname = $"+fmt.Sprint(len(args)+1))
		args = append(args, v)
	}

	if v := filterAndPagination.Address; v != "" {
		where = append(where, "address = $"+fmt.Sprint(len(args)+1))
		args = append(args, v)
	}

	query := `SELECT id, passport_number, passport_series, name, surname, address FROM users`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, filterAndPagination.PerPage, filterAndPagination.PerPage)

	users := make([]*entity.UserDB, 0)
	err := r.conn.SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) DeleteUserById(ctx context.Context, id string) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `DELETE FROM users WHERE id=$1`
	res, err := r.conn.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowAffected == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}

	if tx.Commit() != nil {
		return err
	}
	return nil
}

func (r *UserRepository) DeleteUserByPassport(ctx context.Context, passport string) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `DELETE FROM users WHERE passport_number=$1 AND passport_series=$2`

	passportData := strings.Split(passport, " ")
	passportSeries := passportData[0]
	passportNumber := passportData[1]

	res, err := r.conn.ExecContext(ctx, query, passportNumber, passportSeries)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		//TODO: add error as variable
		return fmt.Errorf("user with passport %s not found", passport)
	}

	if tx.Commit() != nil {
		return err
	}

	return nil
}

func (r *UserRepository) UpdateUserById(ctx context.Context, id string, user *model.User) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	setClause := []string{}
	args := map[string]interface{}{}

	if user.Name != "" {
		setClause = append(setClause, "name=:name")
		args["name"] = user.Name
	}

	if user.Surname != "" {
		setClause = append(setClause, "surname=:surname")
		args["surname"] = user.Surname
	}

	if user.Address != "" {
		setClause = append(setClause, "address=:address")
		args["address"] = user.Address
	}

	if len(setClause) == 0 {
		return errors.New("no fields to update")
	}

	args["id"] = id

	query := fmt.Sprintf("UPDATE users SET %s WHERE id=:id", strings.Join(setClause, ", "))

	res, err := r.conn.NamedExecContext(ctx, query, args)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		return fmt.Errorf("user with id %s not found", id)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) UpdateUserByPassport(ctx context.Context, passport string, user *model.User) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}
	setClause := []string{}
	args := map[string]interface{}{}

	if user.Name != "" {
		setClause = append(setClause, "name=:name")
		args["name"] = user.Name
	}

	if user.Surname != "" {
		setClause = append(setClause, "surname=:surname")
		args["surname"] = user.Surname
	}

	if user.Address != "" {
		setClause = append(setClause, "address=:address")
		args["address"] = user.Address
	}

	if len(setClause) == 0 {
		//TODO: add error as variable
		return errors.New("no fields to update")
	}

	passportData := strings.Split(passport, " ")
	args["passport_series"] = passportData[0]
	args["passport_number"] = passportData[1]

	query := fmt.Sprintf("UPDATE users SET %s WHERE passport_number=:passport_number AND passport_series=:passport_series", strings.Join(setClause, ", "))

	res, err := r.conn.NamedExecContext(ctx, query, args)
	if err != nil {
		return err
	}

	rowAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		//TODO: add error as variable
		return fmt.Errorf("user with passport %s not found", passport)
	}

	if tx.Commit() != nil {
		return err
	}

	return nil
}
