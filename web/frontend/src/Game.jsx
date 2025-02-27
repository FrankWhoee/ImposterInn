import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';

function Game({ws}) {
  return (
    <>
      <h1>Liar's Fortress</h1>
      <h1>Game page hehe</h1>

      <Stack spacing={2} alignItems="center">
        <TextField id="outlined-basic" label="Username" variant="outlined"/>
        <Stack spacing={2} direction="row" alignItems="justify">
          <Button variant="contained">Join Lobby</Button>
          <Button variant="contained">Create Lobby</Button>
        </Stack>
      </Stack>

    </>
  )
}

export default Game