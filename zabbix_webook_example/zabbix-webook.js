try {

    Zabbix.log(4, "[Webhook] Raw: " + value);

    var rawReq = JSON.parse(value);

    var req = new HttpRequest();
    req.addHeader('Content-Type: application/json');

    var body = JSON.stringify({
      status:       rawReq.status,
      severity:     rawReq.severity,
      host:         rawReq.host,
      event_id:     rawReq.eventId,
      trigger_name: rawReq.eventName,
      message:      rawReq.message,
      secret:       rawReq.ZbxNotifierKey
    });

    Zabbix.log(4, "[Webhook] Body: " + body);

    var response = req.post('https://' + rawReq.zabbixWebHost + '/zbx_telegram_notifier/', body);
    var status = req.getStatus();

    if (status < 200 || status >= 300) {
        Zabbix.Log(4, '[Telegram Webhook] notification failed: ' + response);
        throw 'HTTP request failed with status: ' + status + ', response: ' + response;
    }

    Zabbix.Log(4, '[Telegram Webhook] notification ok: ' + status);
    return 'OK';
}
catch (error) {
    Zabbix.Log(4, '[Telegram Webhook] notification failed: ' + error);
    throw 'Sending failed: ' + error + '.';
}