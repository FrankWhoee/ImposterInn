import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import TextField from '@mui/material/TextField';
import Stack from '@mui/material/Stack';
import Button from '@mui/material/Button';

function App() {
  const [count, setCount] = useState(0)

  return (
    <>
      <h1>Imposter Inn</h1>

      <Stack spacing={2}>
      <TextField id="outlined-basic" label="Lobby ID" variant="outlined" />
      <Stack spacing={2} direction="row">
        <Button variant="contained">Join Lobby</Button>
        <Button variant="contained">Create Lobby</Button>
      </Stack>
      </Stack>
      
    </>
  )
}

export default App
