try {
    var params = JSON.parse(value);

    var zbx_subject = params.Subject;
    var zbx_message = params.Message;
    var zbx_notifier_key = params.ZbxNotifierKey;

    if (!zbx_subject || !zbx_notifier_key) {
        throw 'Missing required parameter: Message';
    }

    var zsbj = JSON.parse(zbx_subject)

    var req = new HttpRequest();
    req.addHeader('Content-Type: application/json');

    var body = JSON.stringify({
      status:       zsbj.status,
      trigger_id:   zsbj.triggerId,
      severity:     zsbj.severity,
      host:         zsbj.host,
      event_id:     zsbj.eventId,
      trigger_name: zsbj.eventName,
      message:      zbx_message,
      secret:       zbx_notifier_key
    });

    var response = req.post('https://' + window.location.host + '/zbx_telegram_notifier/', body);
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