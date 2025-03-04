import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';

function LobbyJoin({ ws, setPage, setLobbyId, webId }) {

  let [error, setError] = useState(false)
  let [errorMessage, setErrorMessage] = useState("")

  function goBack() {
    console.log("go back")
    setPage("username")
  }

  async function handleJoinLobby() {
    let lobbyId = document.getElementById("lobbyId").value
    let response = await fetch("/lbjn", { method: "POST", body: JSON.stringify({ webId: webId, lobbyId: lobbyId }) })
    if (!response.ok) {
      let data = await response.json()
      setError(true)
      setErrorMessage(data.message)
    } else {
      let data = await response.json()
      localStorage.setItem("lobbyId", data.lobbyId)
      setLobbyId(data.lobbyId)
      setPage("lobby")
    }
  }

  async function handleCreateLobby() {
    let response = await fetch("/lbcr", { method: "POST", body: JSON.stringify({ webId: webId }) })
    if (!response.ok) {
      let data = await response.json()
      setError(true)
      setErrorMessage(data.message)
    } else {
      let data = await response.json()
      localStorage.setItem("lobbyId", data.lobbyId)
      setLobbyId(data.lobbyId)
      setPage("lobby") 
    }
  }



  return (
    <>
      <h1>Liar's Fortress</h1>

      <Stack spacing={2} alignItems="center">
        <TextField id="lobbyId" label="Lobby ID" variant="outlined" error={error} helperText={errorMessage}/>
        <Stack spacing={2} direction="row" alignItems="justify">
          <Button variant="contained" onClick={handleJoinLobby}>Join Lobby</Button>
          <Button variant="contained" onClick={handleCreateLobby}>Create Lobby</Button>
        </Stack>
        <Button variant="contained" onClick={goBack}>
          <ArrowBackIcon />
        </Button>
      </Stack>

    </>
  )
}

export default LobbyJoin