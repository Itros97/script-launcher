# Primera instalación (macOS y Fedora)

---

Esta guía es para dejar listo un ordenador recién sacado de la caja: instalar lo necesario, conectar
la cámara RealSense D405 y hacer la primera medida con Script Launcher. No hace falta saber
programar. Donde haya que escribir algo en la terminal, cópialo y pégalo tal cual.

Sigue solo los pasos de tu sistema: **macOS** (un Mac) o **Fedora** (Linux).

## Qué necesitas

- El ordenador, con conexión a internet.
- La cámara **Intel RealSense D405** y su cable USB-C. Conéctala directamente al ordenador, sin
  hubs ni alargadores.
- La carpeta del proyecto **calibre3d** (te la pasarán en un USB o en una descarga). Es la carpeta
  que contiene `prueba.sh` y otra carpeta llamada `calibre3d`.
- El tapete impreso y la cámara montada a unos 133 mm sobre él (lo explica el `README.md` del
  proyecto).

## 1. Python

No hay que instalarlo a mano. Script Launcher usa Python 3.12 para hablar con la cámara y hacer las
medidas, y lo instala él mismo con un botón (paso 5). No pide contraseña ni abre ningún instalador.

Si en el Mac aparece una ventana que ofrece instalar las «herramientas de desarrollo», ciérrala con
**Cancelar**: no hacen falta.

## 2. Dar permiso para usar la cámara

**Fedora**

Linux no deja usar la cámara a un usuario normal hasta que se le dice que es de confianza. Abre la
aplicación **Terminal** (búscala en el menú), pega estas tres líneas y pulsa Intro. Te pedirá tu
contraseña:

```bash
sudo curl -L -o /etc/udev/rules.d/99-realsense-libusb.rules https://raw.githubusercontent.com/IntelRealSense/librealsense/master/config/99-realsense-libusb.rules
sudo udevadm control --reload-rules
sudo udevadm trigger
```

Después **desconecta la cámara y vuelve a conectarla**.

**macOS**

No hay que instalar nada. Si al ver la cámara aparece un error de permisos USB, mira
[Si algo falla](#si-algo-falla).

## 3. Descargar Script Launcher

Entra en la página de descargas:
**[github.com/Itros97/script-launcher/releases/latest](https://github.com/Itros97/script-launcher/releases/latest)**

**macOS**

1. Descarga **ScriptLauncher-macOS.zip**. Sirve para todos los Mac (M1–M4 e Intel).
2. Abre el zip: aparece **Script Launcher**. Arrástralo a la carpeta **Aplicaciones**.
3. La primera vez, el Mac dirá que no puede comprobar el desarrollador. Es normal: la app no está
   firmada con una cuenta de pago de Apple. Para abrirla:
   - Abre **Ajustes del Sistema → Privacidad y seguridad**.
   - Baja hasta el aviso sobre «Script Launcher» y pulsa **Abrir igualmente**.
   - Confirma con **Abrir**. Solo hay que hacerlo una vez.

**Fedora**

1. Descarga **script-launcher-linux-x86_64.tar.gz**.
2. Abre la carpeta Descargas, haz clic derecho en el archivo y elige **Extraer aquí**.
3. Abre la **Terminal** y pega:

   ```bash
   bash ~/Descargas/script-launcher-linux-x86_64/install.sh
   ```

   Si tu sistema está en inglés, la carpeta se llama `Downloads` en vez de `Descargas`.
4. Script Launcher ya aparece en el menú de aplicaciones.

## 4. Poner el proyecto calibre3d en su sitio

Copia la carpeta **calibre3d** a tu carpeta **Documentos** (en Fedora, `Documentos`; en Mac,
`Documentos` también). Así no se pierde entre las descargas.

Dentro de esa carpeta deben verse, al menos, `prueba.sh` y otra carpeta `calibre3d`. No hace falta
ningún `requirements.txt`: Script Launcher pone el suyo al preparar el entorno.

## 5. Preparar Calibre3D en Script Launcher

1. Abre **Script Launcher**.
2. Arriba a la izquierda, en el desplegable, elige **Calibre3D**.
3. En la tarjeta **Python**, pulsa el botón azul **Instalar Python 3.12**. Tarda un minuto (descarga
   unos 25 MB en Mac y 67 MB en Linux). Cuando termine dice 🟢 **Python 3.12 listo**. Solo hay que
   hacerlo una vez en cada ordenador.
4. En la tarjeta **Proyecto**, pulsa **Elegir carpeta**. Se abre un selector (sus botones están en
   inglés): pulsa **Home** a la izquierda, haz clic en **Documentos** y luego en **calibre3d** (la
   de fuera, la que tiene `prueba.sh`) para entrar en ella, y pulsa **Choose**.
5. En la tarjeta **Entorno**, pulsa **Preparar entorno**. A la derecha verás cómo se instala todo.
   Tarda unos minutos y descarga unos 300 MB.
6. Cuando termine, el entorno dice 🟢 **Listo**.

Siempre que no sepas qué hacer, pulsa el botón azul: es el siguiente paso.

## 6. Comprobar que todo funciona

1. En la tarjeta **Medir**, pulsa **Demo sin cámara**. En unos segundos aparecen arriba las medidas
   de una pieza simulada (unos 60 × 40 × 20 mm) y, debajo, su mapa de calor. Si lo ves, el
   programa funciona.
2. Con la D405 conectada, en la tarjeta **Cámara** pulsa **Ver cámara**. A la derecha debe verse lo
   que ve la cámara. Pulsa **Parar cámara** cuando acabes.

<p align="center">
  <img alt="Script Launcher en el perfil Calibre3D tras la demo: arriba el resultado con largo 59,97 mm, ancho 40,05 mm, espesor 19,99 mm y Feret 60,04 mm; debajo la salida del programa, el mapa de calor y la vista de la pieza" src="images/calibre3d-medida.png" width="720">
</p>

## 7. Medir una pieza

Calienta la cámara unos 10 minutos antes de empezar: la medida cambia un poco con la temperatura.

1. **1. Plano**: deja el tapete vacío y pulsa el botón. Se hace una vez al empezar la sesión.
2. **2. Centrar**: pon una moneda sobre la cruz del tapete y pulsa. A la derecha dice cuánto mover
   el tapete; muévelo y repite hasta que diga «Centrado».
3. **3. Medir**: pon la pieza dentro del círculo, pulsa, escribe un nombre (por ejemplo
   `muestra_001`, sin espacios) y pulsa **Guardar**. Arriba aparecen el largo, el ancho, el espesor
   y el Feret; debajo, el mapa de calor.

Todo se guarda en la carpeta del proyecto, en `datos/resultados/`. Las fotos de **Guardar imagen**
van a `datos/fotos/`.

## Si algo falla

| Lo que ves | Qué hacer |
|---|---|
| «La descarga de Python está dañada» | Se cortó la conexión a mitad. Pulsa **Instalar Python 3.12** otra vez. |
| «No encuentro python3 y no puedo descargarlo» | Tu ordenador no es ni un Mac ni un Linux normal de 64 bits. Instala Python 3.12 desde [python.org](https://www.python.org/downloads/) y vuelve a abrir Script Launcher. |
| «Sin proyecto» o «En esta carpeta no está el programa calibre3d» | Pulsa **Elegir carpeta**, entra en la carpeta calibre3d de fuera (la que tiene `prueba.sh`) y pulsa **Choose**. |
| «Esta carpeta tiene prueba.sh pero le falta la carpeta calibre3d/» | Falta parte del proyecto: vuelve a copiar la carpeta completa. |
| «Incompleto» con «Falta la librería…» | Pulsa **Preparar entorno** otra vez. Si se repite, añade la librería que dice al `requirements.txt` que Script Launcher dejó en la carpeta del proyecto. |
| La instalación falla con errores de red | Comprueba la conexión a internet y pulsa **Preparar entorno** otra vez. |
| «No se encuentra ninguna RealSense conectada» | Conecta la cámara directamente al ordenador, sin hub. En Fedora, repite el paso 2 y reconecta la cámara. |
| En Mac, un error de permisos USB al usar la cámara | Por verificar en un Mac: en algunas versiones de macOS la librería de la cámara necesita permisos de administrador. Avisa a quien te pasó el programa. |
| En Mac, la instalación dice que no encuentra `pyrealsense2-macosx` | Esa librería necesita macOS 15 o posterior: actualiza el Mac en **Ajustes del Sistema → General → Actualización de software** y pulsa **Preparar entorno** otra vez. |
| En un Mac con Intel, «Ver cámara» dice que no está instalado pyrealsense2 | La librería de la cámara no existe para Mac con Intel: ahí la cámara no funciona. La **Demo sin cámara** sí. |
| La medida da avisos en naranja («Pieza lejos del centro»…) | Lee el aviso: suele bastar con centrar la pieza o repetir el plano. |

> **Nota:** la parte de macOS no se ha probado todavía en un Mac con la D405. Si algo no coincide
> con lo que ves, apúntalo y pásalo para corregir la guía.
