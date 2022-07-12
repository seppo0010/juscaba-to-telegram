CREATE TABLE IF NOT EXISTS expediente (
  numero INT NOT NULL,
  anio INT NOT NULL,
  radicacion_secretaria_primera_instancia VARCHAR(255) NOT NULL,
  radicacion_organismo_primera_instancia VARCHAR(255) NOT NULL,
  radicacion_secretaria_segunda_instancia VARCHAR(255) NOT NULL,
  radicacion_organismo_segunda_instancia VARCHAR(255) NOT NULL,
  ubicacion_organismo VARCHAR(255) NOT NULL,
  ubicacion_dependencia VARCHAR(255) NOT NULL,
  fecha_inicio TIMESTAMP NOT NULL,
  ultimo_movimiento TIMESTAMP NOT NULL,
  caratula VARCHAR(255) NOT NULL,
  UNIQUE (numero, anio)
);
