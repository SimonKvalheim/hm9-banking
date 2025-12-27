import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { getAccounts, getAccountBalance, createTransfer } from '@/api/client';
import type { Account, AccountBalance } from '@/types';
import { ArrowLeft, CheckCircle2 } from 'lucide-react';

interface AccountWithBalance extends Account {
  balance?: AccountBalance;
}

export function TransferPage() {
  const navigate = useNavigate();
  const [accounts, setAccounts] = useState<AccountWithBalance[]>([]);
  const [fromAccountId, setFromAccountId] = useState('');
  const [toAccountId, setToAccountId] = useState('');
  const [amount, setAmount] = useState('');
  const [reference, setReference] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const fetchAccounts = async () => {
      try {
        const accountList = await getAccounts();

        const accountsWithBalances = await Promise.all(
          accountList.map(async (account) => {
            try {
              const balance = await getAccountBalance(account.id);
              return { ...account, balance };
            } catch {
              return account;
            }
          })
        );

        setAccounts(accountsWithBalances);
        if (accountsWithBalances.length > 0) {
          setFromAccountId(accountsWithBalances[0].id);
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to fetch accounts');
      } finally {
        setIsLoading(false);
      }
    };

    fetchAccounts();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const selectedAccount = accounts.find(a => a.id === fromAccountId);
      const currency = selectedAccount?.currency || 'NOK';

      await createTransfer(fromAccountId, toAccountId, amount, currency, reference || undefined);
      setSuccess(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Transfer failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatCurrency = (amount: string, currency: string) => {
    const num = parseFloat(amount);
    return new Intl.NumberFormat('nb-NO', {
      style: 'currency',
      currency: currency,
    }).format(num);
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md text-center">
          <CardHeader>
            <div className="mx-auto mb-4">
              <CheckCircle2 className="h-16 w-16 text-green-500" />
            </div>
            <CardTitle>Transfer Successful</CardTitle>
            <CardDescription>
              Your transfer of {amount} {accounts.find(a => a.id === fromAccountId)?.currency || 'NOK'} has been submitted.
            </CardDescription>
          </CardHeader>
          <CardFooter className="flex justify-center">
            <Button onClick={() => navigate('/dashboard')}>
              Back to Dashboard
            </Button>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b">
        <div className="container mx-auto px-4 py-4">
          <Button variant="ghost" size="sm" asChild>
            <Link to="/dashboard">
              <ArrowLeft className="h-4 w-4 mr-2" />
              Back to Dashboard
            </Link>
          </Button>
        </div>
      </header>

      {/* Main Content */}
      <main className="container mx-auto px-4 py-8 max-w-md">
        <Card>
          <CardHeader>
            <CardTitle>Make a Transfer</CardTitle>
            <CardDescription>
              Transfer money between accounts
            </CardDescription>
          </CardHeader>
          <form onSubmit={handleSubmit}>
            <CardContent className="space-y-4">
              {error && (
                <div className="p-3 text-sm text-destructive bg-destructive/10 rounded-md">
                  {error}
                </div>
              )}

              <div className="space-y-2">
                <Label htmlFor="fromAccount">From Account</Label>
                <select
                  id="fromAccount"
                  className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  value={fromAccountId}
                  onChange={(e) => setFromAccountId(e.target.value)}
                  required
                >
                  {accounts.map((account) => (
                    <option key={account.id} value={account.id}>
                      {account.account_type} - {account.account_number}
                      {account.balance && ` (${formatCurrency(account.balance.available_balance, account.balance.currency)})`}
                    </option>
                  ))}
                </select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="toAccount">To Account ID</Label>
                <Input
                  id="toAccount"
                  placeholder="Enter destination account ID"
                  value={toAccountId}
                  onChange={(e) => setToAccountId(e.target.value)}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  Enter the UUID of the destination account
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="amount">Amount</Label>
                <Input
                  id="amount"
                  type="number"
                  step="0.01"
                  min="0.01"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  required
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="reference">Reference (optional)</Label>
                <Input
                  id="reference"
                  placeholder="Payment reference"
                  value={reference}
                  onChange={(e) => setReference(e.target.value)}
                  maxLength={100}
                />
              </div>
            </CardContent>
            <CardFooter>
              <Button type="submit" className="w-full" disabled={isSubmitting || accounts.length === 0}>
                {isSubmitting ? 'Processing...' : 'Transfer'}
              </Button>
            </CardFooter>
          </form>
        </Card>

        {accounts.length === 0 && (
          <p className="mt-4 text-center text-sm text-muted-foreground">
            You need at least one account to make transfers.
          </p>
        )}
      </main>
    </div>
  );
}
