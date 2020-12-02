<?php

$database_server = 'localhost';
$database_user = 'priornkf_empaci';
$database_password = '6jiTGxOX';
$dbase = 'priornkf_empaci';

$api_key = "VqZK2vrevJYFsv4ZgqEr8tBzr";
$api_secret = "1td3YWNeUvZmu8vqrcLVXJSIBV1JHPV4kzITb6eNHWKjYuFXPj";
$api_auth_url = "https://api.twitter.com/oauth2/token";
$env = "prod";

$tableName = $table_prefix . 'twits';

$api_search_url = "https://api.twitter.com/1.1/tweets/search/30day/prod.json";

$token = "";
$db = false;

$db = initDB($database_server, $database_user, $database_password, $dbase);

function initDB($host, $user, $pwd, $database)
{
    $link = mysqli_connect(
        $host,  /* Хост, к которому мы подключаемся */
        $user,       /* Имя пользователя */
        $pwd,   /* Используемый пароль */
        $database);     /* База данных для запросов по умолчанию */

    if (!$link) {
        printf("Невозможно подключиться к базе данных. Код ошибки: %s\n", mysqli_connect_error());
        exit;
    }

    return $link;
}

function selectTwits($db, $tName)
{
    $res = [];
    if ($result = mysqli_query($db, 'SELECT id, data, created_at FROM ' . $tName)) {

        /* Выборка результатов запроса */
        while ($row = mysqli_fetch_assoc($result)) {
            $res[] = $row;
        }

        /* Освобождаем используемую память */
        mysqli_free_result($result);
    }

    return $res;
}

function insertTwits($db, $tName, $data)
{
    $stmt = $db->prepare("INSERT INTO " . $tName . " (data) VALUES (?)");
    $stmt->bind_param($data);

    /* выполнение подготовленного выражения  */
    $stmt->execute();

    /* Закрытие соединения и выражения*/
    $stmt->close();
}

function getTwits($url, $token, $query)
{
    if ($curl = curl_init()) {
        curl_setopt($curl, CURLOPT_HTTPHEADER, array(
            'Authorization: Bearer ' . $token
        ));
        curl_setopt($curl, CURLOPT_URL, $url);
        curl_setopt($curl, CURLOPT_POST, true);
        curl_setopt($curl, CURLOPT_POSTFIELDS, $query);
        curl_setopt($curl, CURLOPT_RETURNTRANSFER, true);
        $out = curl_exec($curl);
        $jout = json_decode($out);
        curl_close($curl);

        return $jout;
    }

    return false;
}

function getQuery($q, $from = "", $to = "", $next = "")
{
    $query = '{';
    $query .= '"query": "' . $q . '""';

    if (!empty($from)) {
        $query .= ', "fromDate": "' . $from . '"';
    }

    if (!empty($to)) {
        $query .= ', "toDate": "' . $to . '"';
    }

    if (!empty($next)) {
        $query .= ', "next": "' . $next . '"';
    }

    return $query;
}


if ($curl = curl_init()) {
    curl_setopt($curl, CURLOPT_URL, $api_auth_url);
    curl_setopt($curl, CURLOPT_USERPWD, "$api_key:$api_secret");
    curl_setopt($curl, CURLOPT_POST, true);
    curl_setopt($curl, CURLOPT_POSTFIELDS, "grant_type=client_credentials");
    curl_setopt($curl, CURLOPT_RETURNTRANSFER, true);
    $out = curl_exec($curl);
    $jout = json_decode($out);
    if (isset($jout->access_token)) {
        $token = $jout->access_token;
    } else {
        print_r("no access token");
        exit(1);
    }
    curl_close($curl);
}
sleep(1);

$results = [];

if (!empty($token)) {
    $q = "#empaci";
    $query = getQuery($q);
    $isInit = true;

    $now = new DateTime();
    $yesterday = $now->sub(new DateInterval('P1D'));
    $fromDate = $yesterday->format("Ymd") . '0000';
    $toDate = date("Ymd") . '0000';


    $dbTwits = selectTwits($db, $tableName);
    if (!empty($dbTwits)) {
        $isInit = false;
        $query = getQuery($q, $fromDate, $toDate);
    }

    $jout = getTwits($api_search_url, $token, $query);
    if (!empty($jout->results)) {
        $results = $jout->results;
    }

    sleep(1);

    if (isset($jout->next) && !empty($jout->next)) {
        while (true) {
            if ($isInit) {
                $query = getQuery($q, "", "", $jout->next);
            } else {
                $query = getQuery($q, $fromDate, $toDate, $jout->next);
            }

            $jout = getTwits($api_search_url, $token, $query);
            if (!empty($jout->results)) {
                $results = $jout->results;
            }

            if (!isset($jout->next) || empty($jout->next)) {
                break;
            }

            sleep(2);
        }
    }
}

if (!empty($results)) {
    foreach ($results as $r) {
        insertTwits($db, $tableName, json_encode($r));
    }
}


mysqli_close($db);

