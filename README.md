# joss_notify

Plugin oficial para habilitar el envío de notificaciones Push y del sistema (`Notify`) en el lenguaje Joss.

## Instalación

```bash
joss pub add joss_notify
```

## Uso

```joss
use joss_notify;

Notify::title("Aviso")->message("Hola Mundo")->send();
```
