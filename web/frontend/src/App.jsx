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
import { ThemeProvider, createTheme } from '@mui/material/styles';

function App({ ws }) {
  var localStorageUsername = localStorage.getItem("LFusername");
  var localStorageLobbyId = localStorage.getItem("LFlobbyId");
  var localStorageWebId = localStorage.getItem("LFwebId");

  console.log(localStorageUsername);
  console.log(localStorageLobbyId);
  console.log(localStorageWebId);

  var page0
  if (localStorageUsername == null) {
    page0 = "username"
  }
  else if (localStorageLobbyId == null) {
    page0 = "lobbyjoin"
  }
  else {
    page0 = "game"
  }

  const [username, setUsername] = useState(localStorageUsername);
  const [lobbyId, setLobbyId] = useState(localStorageLobbyId);
  const [webId, setWebId] = useState(localStorageWebId);
  const [page, setPage] = useState(page0);

  console.log(page)

  ws.onopen = function () {
    console.log("Connected to server");
    if (webId == null) {
      ws.send("rqid");
    } else {
      ws.send(`amid ${webId}`);
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
    setWebId(args[0])
    localStorage.setItem("LFwebId", args[0])
  }

  const theme = createTheme({
    palette: {
      mode: 'dark',
    },
  });

  return (
    <ThemeProvider theme={theme}>
      {page}
      {page == "username" && <UsernameRegister ws={ws} setUsername={setUsername} username={username} setPage={setPage} />}
      {page == "lobbyjoin" && <LobbyJoin ws={ws} setPage={setPage} setLobbyId={setLobbyId} webId={webId}/>}
      {page == "lobby" && <Lobby ws={ws} setPage={setPage} lobbyId={lobbyId}/>}
      {page == "game" && <Game ws={ws} />}
    </ThemeProvider>
  )
}

export default App