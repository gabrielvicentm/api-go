# Push Notifications

## Backend

O backend aceita tokens gerados por `expo-notifications` e envia push pela API HTTP da Expo quando uma notificacao e criada em `/internal/notificacoes`.

Variaveis opcionais:

- `EXPO_PUSH_ENDPOINT`: sobrescreve o endpoint da Expo. Padrao: `https://exp.host/--/api/v2/push/send`.
- `EXPO_ACCESS_TOKEN`: usado como `Bearer` quando o projeto Expo exigir push security.

Execute a migracao em bancos existentes:

```sql
\i migrations/001_push_tokens.sql
```

Novas rotas autenticadas:

```http
POST /admin/notificacoes/push-token
POST /motorista/notificacoes/push-token
DELETE /admin/notificacoes/push-token
DELETE /motorista/notificacoes/push-token
```

Cadastro:

```json
{
  "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
  "platform": "android",
  "device_id": "device-installation-id"
}
```

Remocao:

```json
{
  "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]"
}
```

## React Native com Expo

Instale:

```bash
npx expo install expo-notifications expo-device expo-constants
```

Exemplo de registro depois do login:

```ts
import Constants from "expo-constants";
import * as Device from "expo-device";
import * as Notifications from "expo-notifications";
import { Platform } from "react-native";

export async function registerPushToken(apiBaseUrl: string, accessToken: string) {
  if (!Device.isDevice) return null;

  const current = await Notifications.getPermissionsAsync();
  const finalStatus =
    current.status === "granted"
      ? current.status
      : (await Notifications.requestPermissionsAsync()).status;

  if (finalStatus !== "granted") return null;

  if (Platform.OS === "android") {
    await Notifications.setNotificationChannelAsync("default", {
      name: "default",
      importance: Notifications.AndroidImportance.MAX,
    });
  }

  const projectId =
    Constants.expoConfig?.extra?.eas?.projectId ??
    Constants.easConfig?.projectId;

  const { data: token } = await Notifications.getExpoPushTokenAsync({
    projectId,
  });

  await fetch(`${apiBaseUrl}/motorista/notificacoes/push-token`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      token,
      platform: Platform.OS,
      device_id: Constants.sessionId,
    }),
  });

  return token;
}
```

Para admin, troque a URL para `/admin/notificacoes/push-token`.
