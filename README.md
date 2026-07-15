# joss_notify 2.0

`joss_notify` envia notificaciones push, in-app y a gateways HTTP. Es un JP v2 autocontenido con bytecode Joss, metadatos de IntelliSense y sidecars para Windows, Linux y macOS (amd64 y arm64). Se carga automaticamente, sin `use` ni dependencias de compilacion en el proyecto consumidor.

## Instalacion

```bash
joss pub add joss_notify 2.0.0
```

## Configuracion

Elige uno de estos modos en `env.joss`:

| Variable | Uso |
| --- | --- |
| `NOTIFY_WEBHOOK_URL` | URL del gateway de notificaciones de la aplicacion. |
| `NOTIFY_WEBHOOK_TOKEN` | Token opcional enviado al gateway. |
| `FCM_SERVER_KEY` | Credencial para envio directo compatible con FCM legacy. |

El modo webhook es el indicado cuando la aplicacion persiste notificaciones o las transmite por WebSocket. El modo FCM usa `user(token)` o `segment(topic)`. Si no hay un backend configurado, el plugin devuelve un error explicito; nunca aparenta haber enviado una notificacion.

## Uso

```joss
$result = Notify::title("Aviso")
    ->message("Tu respaldo termino correctamente")
    ->segment("usuarios")
    ->send()
```

La cadena acepta `app($id)` o `apps($ids)`, `segment($name)`, `user($token)`, `title($text)`, `message($text)`, `html($content)`, `inApp()` y `schedule($timestamp)`. `send()` devuelve la respuesta del proveedor para que la aplicacion pueda manejar los errores de forma explicita.

## Distribucion y desarrollo

`joss_notify.jp` declara seis binarios nativos y contiene `META-INF/joss-symbols.json`, por lo que el editor puede mostrar las firmas del plugin tras instalarlo. Para regenerar el JP se requiere Go y Joss 3.6.0 o posterior; el usuario final no necesita esas herramientas.
