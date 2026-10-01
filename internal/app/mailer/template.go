package mailer

const otpHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light only">
<meta name="supported-color-schemes" content="light only">
<title>Orbit verification code</title>
<link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
<style>
@import url('https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap');
body { margin: 0; padding: 0; background: #23252b; }
table { border-collapse: collapse; }
@media only screen and (max-width: 520px) {
  .pad { padding: 14px 8px !important; }
  .bezel { padding: 12px 10px 14px !important; }
  .screen-pad { padding: 14px 10px 16px !important; }
  .digit { width: 34px !important; height: 44px !important; font-size: 18px !important; }
  .gap { width: 4px !important; }
  .title { font-size: 14px !important; }
}
</style>
</head>
<body style="margin:0;padding:0;background:#23252b;">
<div style="display:none;max-height:0;overflow:hidden;opacity:0;font-size:1px;line-height:1px;color:#23252b;">Your Orbit code is {{.Code}}. Press START to continue.</div>
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#23252b" style="background:#23252b;">
<tr>
<td class="pad" align="center" style="padding:36px 14px;">

<table role="presentation" width="480" cellpadding="0" cellspacing="0" border="0" style="width:100%;max-width:480px;background:#c9c5b9;border:4px solid #0d0e10;box-shadow:8px 8px 0 #0d0e10;">

<tr>
<td style="padding:14px 20px 10px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td align="left" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:12px;color:#5c594f;letter-spacing:1px;">PLAYER 1</td>
<td align="right" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:12px;color:#5c594f;letter-spacing:1px;">ORBIT&nbsp;&#9679;&nbsp;VERIFY</td>
</tr>
</table>
</td>
</tr>

<tr>
<td style="padding:0 16px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" bgcolor="#4b4a5e" style="background:#4b4a5e;border:4px solid #2b2b33;">
<tr>
<td class="bezel" style="padding:14px 18px 18px;">

<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td width="14" style="width:14px;"><table role="presentation" cellpadding="0" cellspacing="0" border="0"><tr><td width="8" height="8" bgcolor="#d0284a" style="width:8px;height:8px;background:#d0284a;font-size:0;line-height:0;">&nbsp;</td></tr></table></td>
<td align="left" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:7px;line-height:10px;color:#c9c5b9;letter-spacing:1px;">BATTERY</td>
<td align="right" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:7px;line-height:10px;color:#c9c5b9;letter-spacing:1px;">DOT MATRIX</td>
</tr>
</table>

<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin-top:10px;">
<tr>
<td bgcolor="#9bbc0f" style="background:#9bbc0f;border:4px solid #0f380f;">

<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td bgcolor="#0f380f" style="background:#0f380f;padding:8px 12px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td align="left" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:12px;color:#9bbc0f;">1UP</td>
<td align="center" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:12px;color:#9bbc0f;">STAGE 1-1</td>
<td align="right" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:12px;color:#9bbc0f;">&#9829;&#9829;&#9829;</td>
</tr>
</table>
</td>
</tr>

<tr>
<td class="screen-pad" align="center" style="padding:20px 16px 22px;">

<div class="title" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:16px;line-height:24px;color:#0f380f;font-weight:bold;letter-spacing:1px;">INSERT CODE</div>

<div style="font-size:0;line-height:0;height:12px;">&nbsp;</div>

<div style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:16px;color:#306230;">ENTER THIS 6-DIGIT CODE<br>TO CREATE YOUR ACCOUNT</div>

<div style="font-size:0;line-height:0;height:16px;">&nbsp;</div>

<table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center">
<tr>
{{range $i, $d := .Digits}}{{if $i}}<td class="gap" width="6" style="width:6px;font-size:0;line-height:0;">&nbsp;</td>{{end}}<td class="digit" align="center" bgcolor="#8bac0f" width="42" height="54" style="width:42px;height:54px;background:#8bac0f;border:4px solid #0f380f;font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:24px;line-height:30px;color:#0f380f;font-weight:bold;text-align:center;">{{$d}}</td>{{end}}
</tr>
</table>

<div style="font-size:0;line-height:0;height:16px;">&nbsp;</div>

<div style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:14px;color:#306230;">CODE: {{.Code}}</div>

<div style="font-size:0;line-height:0;height:14px;">&nbsp;</div>

<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr><td style="border-top:4px dashed #306230;font-size:0;line-height:0;height:4px;">&nbsp;</td></tr>
</table>

<div style="font-size:0;line-height:0;height:10px;">&nbsp;</div>

<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>
<td align="left" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:14px;color:#0f380f;">TIME&nbsp;{{.Minutes}}&nbsp;MIN</td>
<td align="right" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:8px;line-height:14px;color:#0f380f;">TRIES&nbsp;{{.Tries}}</td>
</tr>
</table>

<div style="font-size:0;line-height:0;height:16px;">&nbsp;</div>

<div style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:10px;line-height:16px;color:#0f380f;font-weight:bold;">&gt; PRESS START &lt;</div>

<div style="font-size:0;line-height:0;height:14px;">&nbsp;</div>

<div style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:7px;line-height:13px;color:#306230;">DID NOT PLAY? IGNORE THIS MAIL.<br>NEVER SHARE THIS CODE.</div>

</td>
</tr>
</table>

</td>
</tr>
</table>

</td>
</tr>
</table>
</td>
</tr>

<tr>
<td align="center" style="padding:18px 20px 6px;">
<div style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:12px;line-height:16px;color:#2b2b33;font-weight:bold;letter-spacing:2px;">ORBIT<span style="color:#8a1c5c;">-BOY</span></div>
</td>
</tr>

<tr>
<td style="padding:8px 26px 4px;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
<tr>

<td align="left" valign="middle" width="110" style="width:110px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0">
<tr>
<td width="26" height="26" style="width:26px;height:26px;font-size:0;line-height:0;">&nbsp;</td>
<td width="26" height="26" bgcolor="#2b2b33" style="width:26px;height:26px;background:#2b2b33;font-size:0;line-height:0;border-top:2px solid #4b4a5e;border-left:2px solid #4b4a5e;border-right:2px solid #4b4a5e;">&nbsp;</td>
<td width="26" height="26" style="width:26px;height:26px;font-size:0;line-height:0;">&nbsp;</td>
</tr>
<tr>
<td width="26" height="26" bgcolor="#2b2b33" style="width:26px;height:26px;background:#2b2b33;font-size:0;line-height:0;border-top:2px solid #4b4a5e;border-left:2px solid #4b4a5e;border-bottom:2px solid #4b4a5e;">&nbsp;</td>
<td width="26" height="26" bgcolor="#2b2b33" style="width:26px;height:26px;background:#2b2b33;font-size:0;line-height:0;">&nbsp;</td>
<td width="26" height="26" bgcolor="#2b2b33" style="width:26px;height:26px;background:#2b2b33;font-size:0;line-height:0;border-top:2px solid #4b4a5e;border-right:2px solid #4b4a5e;border-bottom:2px solid #4b4a5e;">&nbsp;</td>
</tr>
<tr>
<td width="26" height="26" style="width:26px;height:26px;font-size:0;line-height:0;">&nbsp;</td>
<td width="26" height="26" bgcolor="#2b2b33" style="width:26px;height:26px;background:#2b2b33;font-size:0;line-height:0;border-bottom:2px solid #4b4a5e;border-left:2px solid #4b4a5e;border-right:2px solid #4b4a5e;">&nbsp;</td>
<td width="26" height="26" style="width:26px;height:26px;font-size:0;line-height:0;">&nbsp;</td>
</tr>
</table>
</td>

<td align="center" valign="bottom" style="padding-bottom:6px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center">
<tr>
<td align="center" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:7px;line-height:10px;color:#5c594f;padding:0 8px;">SELECT</td>
<td align="center" style="font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:7px;line-height:10px;color:#5c594f;padding:0 8px;">START</td>
</tr>
<tr>
<td align="center" style="padding:4px 8px 0;"><table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center"><tr><td width="34" height="10" bgcolor="#6b6a78" style="width:34px;height:10px;background:#6b6a78;border:2px solid #2b2b33;font-size:0;line-height:0;">&nbsp;</td></tr></table></td>
<td align="center" style="padding:4px 8px 0;"><table role="presentation" cellpadding="0" cellspacing="0" border="0" align="center"><tr><td width="34" height="10" bgcolor="#6b6a78" style="width:34px;height:10px;background:#6b6a78;border:2px solid #2b2b33;font-size:0;line-height:0;">&nbsp;</td></tr></table></td>
</tr>
</table>
</td>

<td align="right" valign="middle" width="110" style="width:110px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" align="right">
<tr>
<td valign="bottom" style="padding-right:10px;padding-top:26px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" style="border-collapse:separate;"><tr><td width="34" height="34" align="center" bgcolor="#8a1c5c" style="width:34px;height:34px;background:#8a1c5c;border:3px solid #4d0f34;border-radius:50%;font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:9px;line-height:34px;color:#e9d6e0;text-align:center;">B</td></tr></table>
</td>
<td valign="top">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" style="border-collapse:separate;"><tr><td width="34" height="34" align="center" bgcolor="#8a1c5c" style="width:34px;height:34px;background:#8a1c5c;border:3px solid #4d0f34;border-radius:50%;font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:9px;line-height:34px;color:#e9d6e0;text-align:center;">A</td></tr></table>
</td>
</tr>
</table>
</td>

</tr>
</table>
</td>
</tr>

<tr>
<td align="right" style="padding:6px 26px 20px;">
<table role="presentation" cellpadding="0" cellspacing="0" border="0" align="right">
<tr>
<td width="5" height="30" bgcolor="#8f8b7f" style="width:5px;height:30px;background:#8f8b7f;font-size:0;line-height:0;">&nbsp;</td>
<td width="8" style="width:8px;font-size:0;line-height:0;">&nbsp;</td>
<td width="5" height="30" bgcolor="#8f8b7f" style="width:5px;height:30px;background:#8f8b7f;font-size:0;line-height:0;">&nbsp;</td>
<td width="8" style="width:8px;font-size:0;line-height:0;">&nbsp;</td>
<td width="5" height="30" bgcolor="#8f8b7f" style="width:5px;height:30px;background:#8f8b7f;font-size:0;line-height:0;">&nbsp;</td>
<td width="8" style="width:8px;font-size:0;line-height:0;">&nbsp;</td>
<td width="5" height="30" bgcolor="#8f8b7f" style="width:5px;height:30px;background:#8f8b7f;font-size:0;line-height:0;">&nbsp;</td>
</tr>
</table>
</td>
</tr>

</table>

<table role="presentation" width="480" cellpadding="0" cellspacing="0" border="0" style="width:100%;max-width:480px;">
<tr>
<td align="center" style="padding:26px 10px 0;font-family:'Press Start 2P','Courier New',Courier,monospace;font-size:7px;line-height:14px;color:#7c8092;">&copy; ORBIT &middot; FIND YOUR PEOPLE. KEEP YOUR ORBIT.</td>
</tr>
</table>

</td>
</tr>
</table>
</body>
</html>`
