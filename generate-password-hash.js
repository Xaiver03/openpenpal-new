const bcrypt = require('bcrypt');

async function generateHashes() {
  const passwords = ['password', 'secret', 'admin123'];
  const costs = [10, 12];
  
  console.log('🔐 Generating password hashes...\n');
  
  for (const password of passwords) {
    console.log(`Password: "${password}"`);
    for (const cost of costs) {
      const hash = await bcrypt.hash(password, cost);
      console.log(`  Cost ${cost}: ${hash}`);
      
      // Verify it works
      const match = await bcrypt.compare(password, hash);
      console.log(`  Verify: ${match ? '✅' : '❌'}`);
    }
    console.log('');
  }
  
  // Test specific hash from database
  console.log('📊 Testing specific database hash...');
  const dbHash = '$2a$12$MqRxL8T66Ntbe.F6HfhD0eMCIRRfpd5AFnLALRGU66P6ghgQdvv8i';
  console.log(`Hash: ${dbHash}`);
  
  for (const password of passwords) {
    const match = await bcrypt.compare(password, dbHash);
    console.log(`  "${password}": ${match ? '✅ MATCH' : '❌ NO MATCH'}`);
  }
  
  // Try some other common passwords
  console.log('\n🔍 Testing other common passwords...');
  const commonPasswords = ['courier_level1', 'password123', '123456', 'test', 'demo'];
  for (const password of commonPasswords) {
    const match = await bcrypt.compare(password, dbHash);
    if (match) {
      console.log(`  "${password}": ✅ MATCH!`);
    }
  }
}

generateHashes();