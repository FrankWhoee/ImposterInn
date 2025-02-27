import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';

function LobbyJoin({ ws, setPage }) {
  
  function goBack() {
    console.log("go back")
    setPage("username")
  }

  return (
    <>
      <h1>Liar's Fortress</h1>

      <Stack spacing={2} alignItems="center">
        <TextField id="outlined-basic" label="Lobby ID" variant="outlined" />
        <Stack spacing={2} direction="row" alignItems="justify">
          <Button variant="contained">Join Lobby</Button>
          <Button variant="contained">Create Lobby</Button>
        </Stack>
        <Button variant="contained" onClick={goBack}>
          <ArrowBackIcon/>
        </Button>
      </Stack>

    </>
  )
}

export default LobbyJoin