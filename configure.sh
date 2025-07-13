#!/bin/bash

if [ "$#" -lt 2 ]; then
  echo "Uso: $0 <CLAVE> <VALOR>"
  exit 1
fi

CLAVE=$1
VALOR=$2

echo -e "\nModificando archivos de configuracion .json...\n"

find . -type f -name "*.json" | while read -r archivo; do
  echo "Modificando $archivo ..."

  # Reemplaza "CLAVE": "valor_antiguo" → "CLAVE": "VALOR"
  sed -i -E "s|\"$CLAVE\"[[:space:]]*:[[:space:]]*\"[^\"]*\"|\"$CLAVE\": \"$VALOR\"|g" "$archivo"
done

echo -e "\nArchivos .json modificados correctamente.\n"