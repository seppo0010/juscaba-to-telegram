CREATE TABLE IF NOT EXISTS actuacion (
  numero INT NOT NULL,
  anio INT NOT NULL,
  id INT NOT NULL,
  titulo VARCHAR(255) NOT NULL,
  fecha_firma TIMESTAMP NOT NULL,
  UNIQUE (numero, anio, id)
);
