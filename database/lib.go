package database

import (
	"database/sql"
	"embed"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/seppo0010/libjuscaba"
	"github.com/sirupsen/logrus"
)

type PostgresService struct {
	client *sql.DB
}

//go:embed migrations/*
var f embed.FS

func runMigrations(connString string) error {
	d, err := iofs.New(f, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return err
	}
	defer m.Close()
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func NewPostgresService(connString string) (*PostgresService, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			// not including error message as it might contain secrets
		}).Error("Failed to connect to PostgreSQL")
		return nil, err
	}
	err = runMigrations(connString)
	if err != nil {
		logrus.WithFields(logrus.Fields{"error": err}).Error("Failed to run migrations")
		return nil, err
	}
	return &PostgresService{
		client: db,
	}, nil
}

func (db *PostgresService) AddExpediente(exp *libjuscaba.Ficha) error {
	_, err := db.client.Exec(`
	INSERT INTO expediente (
		numero,
		anio,
		radicacion_secretaria_primera_instancia,
		radicacion_organismo_primera_instancia,
		radicacion_secretaria_segunda_instancia,
		radicacion_organismo_segunda_instancia,
		ubicacion_organismo,
		ubicacion_dependencia,
		fecha_inicio,
		ultimo_movimiento,
		caratula
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8,
		$9,
		$10,
		$11
	)
	ON CONFLICT (numero, anio) DO UPDATE SET
		radicacion_secretaria_primera_instancia = $3,
		radicacion_organismo_primera_instancia = $4,
		radicacion_secretaria_segunda_instancia = $5,
		radicacion_organismo_segunda_instancia = $6,
		ubicacion_organismo = $7,
		ubicacion_dependencia = $8,
		fecha_inicio = $9,
		ultimo_movimiento = $10,
		caratula = $11
	`,
		exp.Numero,
		exp.Anio,
		exp.Radicaciones.SecretariaPrimeraInstancia,
		exp.Radicaciones.OrganismoPrimeraInstancia,
		exp.Radicaciones.SecretariaSegundaInstancia,
		exp.Radicaciones.OrganismoSegundaInstancia,
		exp.Ubicacion.Organismo,
		exp.Ubicacion.Dependencia,
		time.Unix(int64(exp.FechaInicio/1000), 0),
		time.Unix(int64(exp.UltimoMovimiento/1000), 0),
		exp.Caratula,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"numero": exp.Numero,
			"anio":   exp.Anio,
			"error":  err.Error(),
		}).Error("failed to save expediente")
		return err
	}
	return nil
}

func (db *PostgresService) HasActuacion(exp *libjuscaba.Ficha, act *libjuscaba.Actuacion) (bool, error) {
	count := 0
	err := db.client.QueryRow(`
		SELECT COUNT(*) FROM actuacion
		WHERE numero = $1 AND anio = $2 AND id = $3
		`,
		exp.Numero,
		exp.Anio,
		act.ActId,
	).Scan(&count)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"numero": exp.Numero,
			"anio":   exp.Anio,
			"id":     act.ActId,
			"error":  err.Error(),
		}).Error("failed to fetch actuacion")
		return false, err
	}
	return count > 0, nil
}

func (db *PostgresService) AddActuacion(exp *libjuscaba.Ficha, act *libjuscaba.Actuacion) error {
	_, err := db.client.Exec(`
	INSERT INTO actuacion (
		numero,
		anio,
		id,
		titulo,
		firmantes,
		fecha_firma
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6
	)
	ON CONFLICT (numero, anio, id) DO UPDATE SET
		titulo = $4,
		firmantes = $5,
		fecha_firma = $6
	`,
		exp.Numero,
		exp.Anio,
		act.ActId,
		act.Titulo,
		act.Firmantes,
		time.Unix(int64(act.FechaFirma/1000), 0),
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"numero": exp.Numero,
			"anio":   exp.Anio,
			"id":     act.ActId,
			"error":  err.Error(),
		}).Error("failed to save actuacion")
		return err
	}
	return nil
}
