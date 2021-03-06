package slashingprotection

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/fileutil"
	"github.com/prysmaticlabs/prysm/validator/accounts/prompt"
	"github.com/prysmaticlabs/prysm/validator/db/kv"
	"github.com/prysmaticlabs/prysm/validator/flags"
	slashingProtectionFormat "github.com/prysmaticlabs/prysm/validator/slashing-protection/local/standard-protection-format"
	"github.com/urfave/cli/v2"
)

// ImportSlashingProtectionCLI reads an input slashing protection EIP-3076
// standard JSON file and attempts to insert its data into our validator DB.
//
// Steps:
// 1. Parse a path to the validator's datadir from the CLI context.
// 2. Open the validator database.
// 3. Read the JSON file from user input.
// 4. Call the function which actually imports the data from
// from the standard slashing protection JSON file into our database.
func ImportSlashingProtectionCLI(cliCtx *cli.Context) error {
	var err error
	dataDir := cliCtx.String(cmd.DataDirFlag.Name)
	if !cliCtx.IsSet(cmd.DataDirFlag.Name) {
		dataDir, err = prompt.InputDirectory(cliCtx, prompt.DataDirDirPromptText, cmd.DataDirFlag)
		if err != nil {
			return err
		}
	}
	valDB, err := kv.NewKVStore(cliCtx.Context, dataDir, make([][48]byte, 0))
	if err != nil {
		return errors.Wrapf(err, "could not access validator database at path: %s", dataDir)
	}
	defer func() {
		if err := valDB.Close(); err != nil {
			log.WithError(err).Errorf("Could not close validator DB")
		}
	}()
	protectionFilePath, err := prompt.InputDirectory(cliCtx, prompt.SlashingProtectionJSONPromptText, flags.SlashingProtectionJSONFileFlag)
	if err != nil {
		return errors.Wrap(err, "could not get slashing protection json file")
	}
	if protectionFilePath == "" {
		return fmt.Errorf(
			"no path to a slashing_protection.json file specified, please retry or "+
				"you can also specify it with the %s flag",
			flags.SlashingProtectionJSONFileFlag.Name,
		)
	}
	enc, err := fileutil.ReadFileAsBytes(protectionFilePath)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(enc)
	if err := slashingProtectionFormat.ImportStandardProtectionJSON(
		cliCtx.Context, valDB, buf,
	); err != nil {
		return err
	}
	log.Info("Slashing protection JSON successfully imported")
	return nil
}
