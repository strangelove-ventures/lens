package chain_registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/strangelove-ventures/lens/internal/asset_list"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/strangelove-ventures/lens/internal/chain_info"
)

type CosmosGithubRegistry struct{}

func (c CosmosGithubRegistry) ListChains() ([]string, error) {
	client := github.NewClient(http.DefaultClient)
	var chains []string

	tree, res, err := client.Git.GetTree(
		context.Background(),
		"cosmos",
		"chain-registry",
		"master",
		false)
	if err != nil || res.StatusCode != 200 {
		return chains, err
	}

	for _, entry := range tree.Entries {
		if *entry.Type == "tree" && !strings.HasPrefix(*entry.Path, ".") {
			chains = append(chains, *entry.Path)
		}
	}
	return chains, nil
}

func (c CosmosGithubRegistry) GetChain(name string) (chain_info.ChainInfo, error) {
	client := github.NewClient(http.DefaultClient)

	chainFileName := fmt.Sprintf("%s/chain.json", name)
	fileContent, _, res, err := client.Repositories.GetContents(
		context.Background(),
		"cosmos",
		"chain-registry",
		chainFileName,
		&github.RepositoryContentGetOptions{})
	if err != nil || res.StatusCode != 200 {
		return chain_info.ChainInfo{}, errors.Wrap(err, fmt.Sprintf("error fetching %s", chainFileName))
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return chain_info.ChainInfo{}, err
	}

	var result chain_info.ChainInfo
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return chain_info.ChainInfo{}, err
	}

	return result, nil
}

func (c CosmosGithubRegistry) GetAssets(name string) (asset_list.AssetList, error) {
	cl := github.NewClient(http.DefaultClient)

	chainFileName := fmt.Sprintf("%s/assetlist.json", name)
	ch, _, res, err := cl.Repositories.GetContents(context.Background(), "cosmos", "chain-registry", chainFileName, &github.RepositoryContentGetOptions{})
	if err != nil || res.StatusCode != 200 {
		return asset_list.AssetList{}, errors.Wrap(err, fmt.Sprintf("error fetching %s", chainFileName))
	}

	f, err := ch.GetContent()
	if err != nil {
		return asset_list.AssetList{}, err
	}

	var assetList asset_list.AssetList
	if err := json.Unmarshal([]byte(f), &assetList); err != nil {
		return asset_list.AssetList{}, err
	}
	return assetList, nil
}

func (c CosmosGithubRegistry) getChainFiles(name string) (rc chain_info.ChainInfo, al asset_list.AssetList, err error) {
	eg := errgroup.Group{}
	eg.Go(func() (err error) {
		rc, err = c.GetChain(name)
		return err
	})
	eg.Go(func() (err error) {
		al, err = c.GetAssets(name)
		return err
	})
	err = eg.Wait()
	return
}

func (c CosmosGithubRegistry) SourceLink() string {
	return "https://github.com/cosmos/chain-registry"
}
