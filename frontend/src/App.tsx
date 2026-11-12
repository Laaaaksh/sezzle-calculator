import styles from './App.module.css';
import { Calculator } from './features/calculator/components/Calculator';
import { ThemeToggle } from './features/theme';

function App() {
  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Calculator</h1>
        <ThemeToggle />
      </header>
      <main>
        <Calculator />
      </main>
    </div>
  );
}

export default App;
