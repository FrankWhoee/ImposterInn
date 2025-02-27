import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import LobbyJoin from './LobbyJoin';
import Lobby from './Lobby';
import Game from './Game';
import UsernameRegister from './UsernameRegister';

function App({ ws }) {
  var localStorageUsername = localStorage.getItem("LFusername");
  var localStorageLobbyid = localStorage.getItem("LFlobbyid");
  var localStorageUserId = localStorage.getItem("LFuserid");

  console.log(localStorageUsername);
  console.log(localStorageLobbyid);
  console.log(localStorageUserId);

  var page0
  if (localStorageUsername == null) {
    page0 = "username"
  }
  else if (localStorageLobbyid == null) {
    page0 = "lobbyjoin"
  }
  else {
    page0 = "game"
  }

  const [username, setUsername] = useState(localStorageUsername);
  const [lobbyid, setLobbyid] = useState(localStorageLobbyid);
  const [userid, setUserId] = useState(localStorageUserId);
  const [page, setPage] = useState(page0);

  console.log(page)

  ws.onopen = function () {
    console.log("Connected to server");
    if (userid == null) {
      ws.send("rqid");
    } else {
      ws.send(`amid ${userid}`);
    }

    if (username != null) {
      ws.send(`name ${username}`);
    }
  }


  const cmdToFn = {
    "asid": asid
  }

  ws.onmessage = function (msg) {
    const cmd = msg.data.substring(0, 4)
    const args = msg.data.substring(5).split(" ")

    cmdToFn[cmd](args)
  }

  function asid(args) {
    setUserId(args[0])
    localStorage.setItem("LFuserid", args[0])
  }

  return (
    <>
      {page}
      {page == "username" && <UsernameRegister ws={ws} setUsername={setUsername} username={username} setPage={setPage} />}
      {page == "lobbyjoin" && <LobbyJoin ws={ws} setPage={setPage} />}
      {page == "lobby" && <Lobby ws={ws} setPage={setPage} />}
      {page == "game" && <Game ws={ws} />}
    </>
  )
}

export default App